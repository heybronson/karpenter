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

package schedcache_test

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clock "k8s.io/utils/clock/testing"

	"sigs.k8s.io/karpenter/pkg/controllers/disruption/schedcache"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/utils/pdb"
)

func TestBuildCapturesInputsAndValidity(t *testing.T) {
	clk := clock.NewFakeClock(time.Unix(1_700_000_000, 0))
	c := schedcache.New(clk)
	if c.Populated() {
		t.Fatal("fresh cache should not be populated")
	}

	snapshot := clk.Now()
	err := c.Build(context.Background(), schedcache.BuildInputs{
		PendingPods:       []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "a"}}},
		PDBs:              pdb.Limits{},
		StateNodes:        []*state.StateNode{},
		ConsolidationMark: snapshot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !c.Populated() {
		t.Fatal("cache should be populated after Build")
	}
	if got := c.PendingPods(); len(got) != 1 || got[0].Name != "a" {
		t.Fatalf("PendingPods = %v", got)
	}

	if !c.IsValidAgainst(snapshot) {
		t.Fatal("cache should be valid when current mark equals snapshot")
	}
	if c.IsValidAgainst(snapshot.Add(time.Minute)) {
		t.Fatal("cache should be invalid when current mark advances past snapshot")
	}
}
