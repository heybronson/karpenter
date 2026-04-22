/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schedcache

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/clock"

	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/utils/pdb"
)

// PassCache holds SimulateScheduling inputs that are stable within one consolidation pass.
type PassCache struct {
	clk               clock.Clock
	built             bool
	pendingPods       []*corev1.Pod
	pdbs              pdb.Limits
	stateNodes        []*state.StateNode
	consolidationMark time.Time
}

// BuildInputs are the stable inputs captured at the start of a pass.
type BuildInputs struct {
	PendingPods       []*corev1.Pod
	PDBs              pdb.Limits
	StateNodes        []*state.StateNode
	ConsolidationMark time.Time
}

// New returns an unpopulated PassCache using the provided clock.
func New(clk clock.Clock) *PassCache { return &PassCache{clk: clk} }

// Build captures the stable inputs for one consolidation pass.
func (c *PassCache) Build(_ context.Context, in BuildInputs) error {
	c.pendingPods = in.PendingPods
	c.pdbs = in.PDBs
	c.stateNodes = in.StateNodes
	c.consolidationMark = in.ConsolidationMark
	c.built = true
	return nil
}

// Populated reports whether Build has been called.
func (c *PassCache) Populated() bool { return c.built }

// PendingPods returns the pending pods captured at Build time.
func (c *PassCache) PendingPods() []*corev1.Pod { return c.pendingPods }

// PDBs returns the PDB limits captured at Build time.
func (c *PassCache) PDBs() pdb.Limits { return c.pdbs }

// StateNodes returns the state nodes captured at Build time.
func (c *PassCache) StateNodes() []*state.StateNode { return c.stateNodes }

// ConsolidationMark returns the cluster consolidation mark captured at Build time.
func (c *PassCache) ConsolidationMark() time.Time { return c.consolidationMark }

// IsValidAgainst reports whether the cache is still consistent with the current cluster
// consolidation mark. Callers pass state.Cluster.ConsolidationState().
func (c *PassCache) IsValidAgainst(currentMark time.Time) bool {
	if !c.built {
		return false
	}
	return !currentMark.After(c.consolidationMark)
}
