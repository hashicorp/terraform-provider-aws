// Copyright 2016 ALRUX Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package levenshtein

import (
	"testing"
)

func Test_Params(t *testing.T) {
	// Different allocations should not be equal.
	if NewParams() == NewParams() {
		t.Errorf(`NewParams() == NewParams()`)
	}

	// Clone should not be equal to original allocation.
	p := NewParams()
	if p == p.Clone() {
		t.Errorf(`p == p.Clone()`)
	}

	// Defaults
	if p.insCost != 1 {
		t.Errorf(`NewParams().insCost == %v, want %v`, p.insCost, 1)
	}
	if p.subCost != 1 {
		t.Errorf(`NewParams().subCost == %v, want %v`, p.subCost, 1)
	}
	if p.delCost != 1 {
		t.Errorf(`NewParams().delCost == %v, want %v`, p.delCost, 1)
	}
	if p.maxCost != 0 {
		t.Errorf(`NewParams().maxCost == %v, want %v`, p.maxCost, 0)
	}
	if p.minScore != 0 {
		t.Errorf(`NewParams().minScore == %v, want %v`, p.minScore, 0)
	}
	if p.bonusPrefix != 4 {
		t.Errorf(`NewParams().bonusPrefix == %v, want %v`, p.bonusPrefix, 4)
	}
	if p.bonusScale != .1 {
		t.Errorf(`NewParams().bonusScale == %v, want %v`, p.bonusScale, .1)
	}
	if p.bonusThreshold != .7 {
		t.Errorf(`NewParams().bonusThreshold == %v, want %v`, p.bonusThreshold, .7)
	}

	// Setters
	if p = NewParams().InsCost(2); p.insCost != 2 {
		t.Errorf(`NewParams().InsCost(2).insCost == %v, want %v`, p.insCost, 2)
	}
	if p = NewParams().InsCost(-2); p.insCost != 1 {
		t.Errorf(`NewParams().InsCost(-2).insCost == %v, want %v`, p.insCost, 1)
	}
	if p = NewParams().SubCost(3); p.subCost != 3 {
		t.Errorf(`NewParams().SubCost(3).subCost == %v, want %v`, p.subCost, 3)
	}
	if p = NewParams().SubCost(-3); p.subCost != 1 {
		t.Errorf(`NewParams().SubCost(-3).subCost == %v, want %v`, p.subCost, 1)
	}
	if p = NewParams().DelCost(5); p.delCost != 5 {
		t.Errorf(`NewParams().DelCost(5).delCost == %v, want %v`, p.delCost, 5)
	}
	if p = NewParams().DelCost(-1); p.delCost != 1 {
		t.Errorf(`NewParams().DelCost(-1).delCost == %v, want %v`, p.delCost, 1)
	}
	if p = NewParams().MaxCost(7); p.maxCost != 7 {
		t.Errorf(`NewParams().MaxCost(7).maxCost == %v, want %v`, p.maxCost, 7)
	}
	if p = NewParams().MaxCost(-5); p.maxCost != 0 {
		t.Errorf(`NewParams().MaxCost(-5).maxCost == %v, want %v`, p.maxCost, 0)
	}
	if p = NewParams().MinScore(.5); p.minScore != .5 {
		t.Errorf(`NewParams().MinScore(.5).minScore == %v, want %v`, p.minScore, .5)
	}
	if p = NewParams().MinScore(3); p.minScore != 3 {
		t.Errorf(`NewParams().MinScore(3).minScore == %v, want %v`, p.minScore, 3)
	}
	if p = NewParams().MinScore(-5); p.minScore != 0 {
		t.Errorf(`NewParams().MinScore(-5).minScore == %v, want %v`, p.minScore, 0)
	}
	if p = NewParams().BonusPrefix(7); p.bonusPrefix != 7 {
		t.Errorf(`NewParams().BonusPrefix(7).bonusPrefix == %v, want %v`, p.bonusPrefix, 7)
	}
	if p = NewParams().BonusPrefix(-5); p.bonusPrefix != 4 {
		t.Errorf(`NewParams().BonusPrefix(-5).bonusPrefix == %v, want %v`, p.bonusPrefix, 4)
	}
	if p = NewParams().BonusScale(.2); p.bonusScale != .2 {
		t.Errorf(`NewParams().BonusScale(.2).bonusScale == %v, want %v`, p.bonusScale, .2)
	}
	if p = NewParams().BonusScale(-.3); p.bonusScale != .1 {
		t.Errorf(`NewParams().BonusScale(-.3).bonusScale == %v, want %v`, p.bonusScale, .1)
	}
	if p = NewParams().BonusScale(7); p.bonusScale != 1/float64(p.bonusPrefix) {
		t.Errorf(`NewParams().BonusScale(7).bonusScale == %v, want %v`, p.bonusScale, 1/float64(p.bonusPrefix))
	}
	if p = NewParams().BonusThreshold(.3); p.bonusThreshold != .3 {
		t.Errorf(`NewParams().BonusThreshold(.3).bonusThreshold == %v, want %v`, p.bonusThreshold, .3)
	}
	if p = NewParams().BonusThreshold(7); p.bonusThreshold != 7 {
		t.Errorf(`NewParams().BonusThreshold(7).bonusThreshold == %v, want %v`, p.bonusThreshold, 7)
	}
	if p = NewParams().BonusThreshold(-7); p.bonusThreshold != .7 {
		t.Errorf(`NewParams().BonusThreshold(-7).bonusThreshold == %v, want %v`, p.bonusThreshold, .7)
	}

	// Cloning nil pointer should initiate with default values
	var p1 *Params
	p2 := p1.Clone()
	if p2.insCost != 1 {
		t.Errorf(`nil.Clone().insCost == %v, want %v`, p2.insCost, 1)
	}
	if p2.subCost != 1 {
		t.Errorf(`nil.Clone().subCost == %v, want %v`, p2.subCost, 1)
	}
	if p2.delCost != 1 {
		t.Errorf(`nil.Clone().delCost == %v, want %v`, p2.delCost, 1)
	}
	if p2.maxCost != 0 {
		t.Errorf(`nil.Clone().maxCost == %v, want %v`, p2.maxCost, 0)
	}
	if p2.minScore != 0 {
		t.Errorf(`nil.Clone().minScore == %v, want %v`, p2.minScore, 0)
	}
	if p2.bonusPrefix != 4 {
		t.Errorf(`nil.Clone().bonusPrefix == %v, want %v`, p2.bonusPrefix, 4)
	}
	if p2.bonusScale != .1 {
		t.Errorf(`nil.Clone().bonusScale == %v, want %v`, p2.bonusScale, .1)
	}
	if p2.bonusThreshold != .7 {
		t.Errorf(`nil.Clone().bonusThreshold == %v, want %v`, p2.bonusThreshold, .7)
	}
}
