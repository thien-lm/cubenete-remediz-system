/*
Copyright 2018 Planet Labs Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing permissions
and limitations under the License.
*/

package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/planetlabs/draino/internal/vcd"
	"github.com/planetlabs/draino/internal/vcd/manipulation"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

const (
	// DefaultDrainBuffer is the default minimum time between node drains.
	DefaultDrainBuffer = 10 * time.Minute

	eventReasonCordonStarting  = "CordonStarting"
	eventReasonCordonSucceeded = "CordonSucceeded"
	eventReasonCordonFailed    = "CordonFailed"

	eventReasonUncordonStarting  = "UncordonStarting"
	eventReasonUncordonSucceeded = "UncordonSucceeded"
	eventReasonUncordonFailed    = "UncordonFailed"

	eventReasonDrainScheduled        = "DrainScheduled"
	eventReasonDrainSchedulingFailed = "DrainSchedulingFailed"
	eventReasonDrainStarting         = "DrainStarting"
	eventReasonDrainSucceeded        = "DrainSucceeded"
	eventReasonDrainFailed           = "DrainFailed"

	tagResultSucceeded = "succeeded"
	tagResultFailed    = "failed"

	drainRetryAnnotationKey   = "draino/drain-retry"
	drainRetryAnnotationValue = "true"

	drainoConditionsAnnotationKey = "draino.planet.com/conditions"
)

// Opencensus measurements.
var (
	MeasureNodesCordoned       = stats.Int64("draino/nodes_cordoned", "Number of nodes cordoned.", stats.UnitDimensionless)
	MeasureNodesUncordoned     = stats.Int64("draino/nodes_uncordoned", "Number of nodes uncordoned.", stats.UnitDimensionless)
	MeasureNodesDrained        = stats.Int64("draino/nodes_drained", "Number of nodes drained.", stats.UnitDimensionless)
	MeasureNodesDrainScheduled = stats.Int64("draino/nodes_drainScheduled", "Number of nodes drain scheduled.", stats.UnitDimensionless)

	TagNodeName, _ = tag.NewKey("node_name")
	TagResult, _   = tag.NewKey("result")
)

// A DrainingResourceEventHandler cordons and drains any added or updated nodes.
type DrainingResourceEventHandler struct {
	logger         *zap.Logger
	cordonDrainer  CordonDrainer
	eventRecorder  record.EventRecorder
	drainScheduler DrainScheduler

	lastDrainScheduledFor time.Time
	buffer                time.Duration

	conditions []SuppliedCondition
	clientSet     kubernetes.Interface
	lock         sync.Mutex
}

// DrainingResourceEventHandlerOption configures an DrainingResourceEventHandler.
type DrainingResourceEventHandlerOption func(d *DrainingResourceEventHandler)

// WithLogger configures a DrainingResourceEventHandler to use the supplied
// logger.
func WithLogger(l *zap.Logger) DrainingResourceEventHandlerOption {
	return func(h *DrainingResourceEventHandler) {
		h.logger = l
	}
}

// WithDrainBuffer configures the minimum time between scheduled drains.
func WithDrainBuffer(d time.Duration) DrainingResourceEventHandlerOption {
	return func(h *DrainingResourceEventHandler) {
		h.buffer = d
	}
}

// WithConditionsFilter configures which conditions should be handled.
func WithConditionsFilter(conditions []string) DrainingResourceEventHandlerOption {
	return func(h *DrainingResourceEventHandler) {
		h.conditions = ParseConditions(conditions)
	}
}

// NewDrainingResourceEventHandler returns a new DrainingResourceEventHandler.
func NewDrainingResourceEventHandler(cs   kubernetes.Interface, d CordonDrainer, e record.EventRecorder, ho ...DrainingResourceEventHandlerOption) *DrainingResourceEventHandler {
	h := &DrainingResourceEventHandler{
		logger:                zap.NewNop(),
		cordonDrainer:         d,
		eventRecorder:         e,
		lastDrainScheduledFor: time.Now(),
		buffer:                DefaultDrainBuffer,
		clientSet: cs,
	}
	for _, o := range ho {
		o(h)
	}
	h.drainScheduler = NewDrainSchedules(d, e, h.buffer, h.logger)
	return h
}

// OnAdd cordons and drains the added node.
func (h *DrainingResourceEventHandler) OnAdd(obj interface{}) {
	n, ok := obj.(*core.Node)
	if !ok {
		return
	}
	h.HandleNode(n)
}

// OnUpdate cordons and drains the updated node.
func (h *DrainingResourceEventHandler) OnUpdate(_, newObj interface{}) {
	h.OnAdd(newObj)
}

// OnDelete does nothing. There's no point cordoning or draining deleted nodes.

func (h *DrainingResourceEventHandler) OnDelete(obj interface{}) {
	n, ok := obj.(*core.Node)
	if !ok {
		d, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		h.drainScheduler.DeleteSchedule(d.Key)
	}

	h.drainScheduler.DeleteSchedule(n.GetName())
}

func (h *DrainingResourceEventHandler) HandleNode(n *core.Node) {
	h.lock.Lock()
	defer h.lock.Unlock()
	log := h.logger.With(zap.String("node", n.GetName()))

	badConditions := h.offendingConditions(n)
	if len(badConditions) == 0 {
		if shouldUncordon(n) {
			h.drainScheduler.DeleteSchedule(n.GetName())
			h.uncordon(n)
		}
		return
	}

	if !h.shouldBeHandled(n) {
		return
	}

	if !n.Spec.Unschedulable {
		log.Info("trying to reboot ")
		//fetch info to access to vmware platform
		log.Info("trying to reboot node")
		goVcloudClient, org, vdc, err := h.createGoVCloudClient()
		if err != nil {
			log.Info("failed to connect to vcd")
			return
		}
		//reboot vm in infra
		err2 := manipulation.RebootVM(goVcloudClient, org, vdc, n.Name)
		if err2 != nil {
			log.Info("failed to handle not ready node")
			return
		}
		isNodeReady := h.checkNodeReadyStatusAfterRepairing(n)
		if isNodeReady {
			log.Info("repair node perform by auto repair controller was ran successfully")
			return 
		}
		log.Info("Node will be drained and replaced")
	}
	// First cordon the node if it is not yet cordonned
	if !n.Spec.Unschedulable {
		h.cordon(n, badConditions)
	}

	// Let's ensure that a drain is scheduled
	hasSChedule, failedDrain := h.drainScheduler.HasSchedule(n.GetName())
	if !hasSChedule {
		h.scheduleDrain(n)
		return
	}

	// Is there a request to retry a failed drain activity. If yes reschedule drain
	if failedDrain && HasDrainRetryAnnotation(n) {
		h.drainScheduler.DeleteSchedule(n.GetName())
		h.scheduleDrain(n)
		return
	}
}

func (h *DrainingResourceEventHandler) offendingConditions(n *core.Node) []SuppliedCondition {
	var conditions []SuppliedCondition
	for _, suppliedCondition := range h.conditions {
		for _, nodeCondition := range n.Status.Conditions {
			if suppliedCondition.Type == nodeCondition.Type &&
				suppliedCondition.Status == nodeCondition.Status &&
				time.Since(nodeCondition.LastTransitionTime.Time) >= suppliedCondition.MinimumDuration {
				conditions = append(conditions, suppliedCondition)
			}
		}
	}
	return conditions
}

func shouldUncordon(n *core.Node) bool {
	if !n.Spec.Unschedulable {
		return false
	}
	previousConditions := parseConditionsFromAnnotation(n)
	if len(previousConditions) == 0 {
		return false
	}
	for _, previousCondition := range previousConditions {
		for _, nodeCondition := range n.Status.Conditions {
			if previousCondition.Type == nodeCondition.Type &&
				previousCondition.Status != nodeCondition.Status &&
				time.Since(nodeCondition.LastTransitionTime.Time) >= previousCondition.MinimumDuration {
				return true
			}
		}
	}
	return false
}

func parseConditionsFromAnnotation(n *core.Node) []SuppliedCondition {
	if n.Annotations == nil {
		return nil
	}
	if n.Annotations[drainoConditionsAnnotationKey] == "" {
		return nil
	}
	rawConditions := strings.Split(n.Annotations[drainoConditionsAnnotationKey], ";")
	return ParseConditions(rawConditions)
}

func (h *DrainingResourceEventHandler) uncordon(n *core.Node) {
	log := h.logger.With(zap.String("node", n.GetName()))
	tags, _ := tag.New(context.Background(), tag.Upsert(TagNodeName, n.GetName())) // nolint:gosec
	nr := &core.ObjectReference{Kind: "Node", Name: n.GetName(), UID: types.UID(n.GetName())}

	log.Debug("Uncordoning")
	h.eventRecorder.Event(nr, core.EventTypeWarning, eventReasonUncordonStarting, "Uncordoning node")
	if err := h.cordonDrainer.Uncordon(n, removeAnnotationMutator); err != nil {
		log.Info("Failed to uncordon", zap.Error(err))
		tags, _ = tag.New(tags, tag.Upsert(TagResult, tagResultFailed)) // nolint:gosec
		stats.Record(tags, MeasureNodesUncordoned.M(1))
		h.eventRecorder.Eventf(nr, core.EventTypeWarning, eventReasonUncordonFailed, "Uncordoning failed: %v", err)
		return
	}
	log.Info("Uncordoned")
	tags, _ = tag.New(tags, tag.Upsert(TagResult, tagResultSucceeded)) // nolint:gosec
	stats.Record(tags, MeasureNodesUncordoned.M(1))
	h.eventRecorder.Event(nr, core.EventTypeWarning, eventReasonUncordonSucceeded, "Uncordoned node")
}

func removeAnnotationMutator(n *core.Node) {
	delete(n.Annotations, drainoConditionsAnnotationKey)
}

func (h *DrainingResourceEventHandler) cordon(n *core.Node, badConditions []SuppliedCondition) {
	log := h.logger.With(zap.String("node", n.GetName()))
	tags, _ := tag.New(context.Background(), tag.Upsert(TagNodeName, n.GetName())) // nolint:gosec
	// Events must be associated with this object reference, rather than the
	// node itself, in order to appear under `kubectl describe node` due to the
	// way that command is implemented.
	// https://github.com/kubernetes/kubernetes/blob/17740a2/pkg/printers/internalversion/describe.go#L2711
	nr := &core.ObjectReference{Kind: "Node", Name: n.GetName(), UID: types.UID(n.GetName())}

	log.Debug("Cordoning")
	h.eventRecorder.Event(nr, core.EventTypeWarning, eventReasonCordonStarting, "Cordoning node")
	if err := h.cordonDrainer.Cordon(n, conditionAnnotationMutator(badConditions)); err != nil {
		log.Info("Failed to cordon", zap.Error(err))
		tags, _ = tag.New(tags, tag.Upsert(TagResult, tagResultFailed)) // nolint:gosec
		stats.Record(tags, MeasureNodesCordoned.M(1))
		h.eventRecorder.Eventf(nr, core.EventTypeWarning, eventReasonCordonFailed, "Cordoning failed: %v", err)
		return
	}
	log.Info("Cordoned")
	tags, _ = tag.New(tags, tag.Upsert(TagResult, tagResultSucceeded)) // nolint:gosec
	stats.Record(tags, MeasureNodesCordoned.M(1))
	h.eventRecorder.Event(nr, core.EventTypeWarning, eventReasonCordonSucceeded, "Cordoned node")
}

func conditionAnnotationMutator(conditions []SuppliedCondition) func(*core.Node) {
	var value []string
	for _, c := range conditions {
		value = append(value, fmt.Sprintf("%v=%v,%v", c.Type, c.Status, c.MinimumDuration))
	}
	return func(n *core.Node) {
		if n.Annotations == nil {
			n.Annotations = make(map[string]string)
		}
		n.Annotations[drainoConditionsAnnotationKey] = strings.Join(value, ";")
	}
}

// drain schedule the draining activity
func (h *DrainingResourceEventHandler) scheduleDrain(n *core.Node) {
	log := h.logger.With(zap.String("node", n.GetName()))
	tags, _ := tag.New(context.Background(), tag.Upsert(TagNodeName, n.GetName())) // nolint:gosec
	nr := &core.ObjectReference{Kind: "Node", Name: n.GetName(), UID: types.UID(n.GetName())}
	log.Debug("Scheduling drain")
	when, err := h.drainScheduler.Schedule(n)
	if err != nil {
		if IsAlreadyScheduledError(err) {
			return
		}
		log.Info("Failed to schedule the drain activity", zap.Error(err))
		tags, _ = tag.New(tags, tag.Upsert(TagResult, tagResultFailed)) // nolint:gosec
		stats.Record(tags, MeasureNodesDrainScheduled.M(1))
		h.eventRecorder.Eventf(nr, core.EventTypeWarning, eventReasonDrainSchedulingFailed, "Drain scheduling failed: %v", err)
		return
	}
	log.Info("Drain scheduled ", zap.Time("after", when))
	tags, _ = tag.New(tags, tag.Upsert(TagResult, tagResultSucceeded)) // nolint:gosec
	stats.Record(tags, MeasureNodesDrainScheduled.M(1))
	h.eventRecorder.Eventf(nr, core.EventTypeWarning, eventReasonDrainScheduled, "Will drain node after %s", when.Format(time.RFC3339Nano))
}

func HasDrainRetryAnnotation(n *core.Node) bool {
	return n.GetAnnotations()[drainRetryAnnotationKey] == drainRetryAnnotationValue
}

// create vcloud client to access infra
func (h *DrainingResourceEventHandler) createGoVCloudClient() (*govcd.VCDClient, string, string, error) {
	kubeClient := h.clientSet
	userName := vcd.GetUserName(kubeClient)
	password := vcd.GetPassword(kubeClient)
	host     := vcd.GetHost(kubeClient)
	org 	:= vcd.GetORG(kubeClient)
	vdc     := vcd.GetVDC(kubeClient)

	goVCloudClientConfig := vcd.Config{
		User: userName,
		Password: password,
		Href: fmt.Sprintf("%v/api", host),
		Org: org,
		VDC: vdc,
	}
	
	goVcloudClient, err := goVCloudClientConfig.NewClient()
	return goVcloudClient, org, vdc, err
}

//check status of rebooted node in cluster 
//this method trying to fetch node status from Kube API server each 10s
func (h *DrainingResourceEventHandler) checkNodeReadyStatusAfterRepairing(n *core.Node) bool {
	logger := h.logger.With(zap.String("node", n.GetName()))
	logger.Info("Rebooted node in infrastructure, waiting for Ready state in kubernetes")
	maxRetry := 5
	//retryCount := 0
	for retryCount := 0; retryCount < maxRetry; retryCount++ {
			nodes, _ := h.clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
		for _, node := range nodes.Items {
			if node.Name == n.Name {
					for _, condition := range node.Status.Conditions {
						if condition.Type == core.NodeReady && condition.Status == core.ConditionTrue {
							logger.Info("node is healthy now, don't need to replace")
							time.Sleep(15*time.Second)
							return true
						}
				}
			}
		}
		logger.Info("can not determine if node is healthy, retry after 10 seconds")
		time.Sleep(15*time.Second)
	}
	return false
}

func (h *DrainingResourceEventHandler) shouldBeHandled(n *core.Node) bool {
	logger := h.logger.With(zap.String("node", n.GetName()))
	waitingTimeForNotReady := 3*time.Minute
    for _, condition := range n.Status.Conditions {
        if condition.Type == core.NodeReady && condition.Status != core.ConditionTrue {
            lastTransitionTime := condition.LastTransitionTime.Time
            if time.Since(lastTransitionTime) >= waitingTimeForNotReady && !n.Spec.Unschedulable{
				logger.Info("node need to be handled")
                return true
            }
        }
    }
	logger.Info("node is not ready but does not need to be handled right now")
	return false
}