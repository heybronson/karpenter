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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clock "k8s.io/utils/clock/testing"

	"sigs.k8s.io/karpenter/pkg/controllers/disruption/schedcache"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/utils/pdb"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SchedCache")
}

var _ = Describe("PassCache", func() {
	var clk *clock.FakeClock

	BeforeEach(func() {
		clk = clock.NewFakeClock(time.Unix(1_700_000_000, 0))
	})

	It("captures inputs and validates against the consolidation mark", func() {
		c := schedcache.New(clk)
		Expect(c.Populated()).To(BeFalse(), "fresh cache should not be populated")

		snapshot := clk.Now()
		Expect(c.Build(context.Background(), schedcache.BuildInputs{
			PendingPods:       []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "a"}}},
			PDBs:              pdb.Limits{},
			StateNodes:        []*state.StateNode{},
			ConsolidationMark: snapshot,
		})).To(Succeed())
		Expect(c.Populated()).To(BeTrue(), "cache should be populated after Build")

		got := c.PendingPods()
		Expect(got).To(HaveLen(1))
		Expect(got[0].Name).To(Equal("a"))

		Expect(c.IsValidAgainst(snapshot)).To(BeTrue(), "valid when mark equals snapshot")
		Expect(c.IsValidAgainst(snapshot.Add(time.Minute))).To(BeFalse(), "invalid when mark advances")
	})
})
