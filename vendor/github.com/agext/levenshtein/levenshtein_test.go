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

type e struct {
	cost, lp, ls int
	sim, match   float64
}

func Test_Metrics(t *testing.T) {
	var (
		cases = []struct {
			s1   string
			s2   string
			desc string
			p    *Params
			exp  e
		}{
			// When the values are the same...
			{"", "", "", nil, e{0, 0, 0, 1, 1}},
			{"1", "1", "", nil, e{0, 1, 0, 1, 1}},
			{"12", "12", "", nil, e{0, 2, 0, 1, 1}},
			{"123", "123", "", nil, e{0, 3, 0, 1, 1}},
			{"1234", "1234", "", nil, e{0, 4, 0, 1, 1}},
			{"12345", "12345", "", nil, e{0, 5, 0, 1, 1}},
			{"password", "password", "", nil, e{0, 8, 0, 1, 1}},

			// When one of the values is empty...
			{"", "1", "", nil, e{1, 0, 0, 0, 0}},
			{"", "12", "", nil, e{2, 0, 0, 0, 0}},
			{"", "123", "", nil, e{3, 0, 0, 0, 0}},
			{"", "1234", "", nil, e{4, 0, 0, 0, 0}},
			{"", "12345", "", nil, e{5, 0, 0, 0, 0}},
			{"", "password", "", nil, e{8, 0, 0, 0, 0}},
			{"1", "", "", nil, e{1, 0, 0, 0, 0}},
			{"12", "", "", nil, e{2, 0, 0, 0, 0}},
			{"123", "", "", nil, e{3, 0, 0, 0, 0}},
			{"1234", "", "", nil, e{4, 0, 0, 0, 0}},
			{"12345", "", "", nil, e{5, 0, 0, 0, 0}},
			{"password", "", "", nil, e{8, 0, 0, 0, 0}},

			// When a single character is inserted or removed...
			{"password", "1password", "", nil, e{1, 0, 8, 8.0 / 9, 8.0 / 9}},
			{"password", "p1assword", "", nil, e{1, 1, 7, 8.0 / 9, 8.1 / 9}},
			{"password", "pa1ssword", "", nil, e{1, 2, 6, 8.0 / 9, 8.2 / 9}},
			{"password", "pas1sword", "", nil, e{1, 3, 5, 8.0 / 9, 8.3 / 9}},
			{"password", "pass1word", "", nil, e{1, 4, 4, 8.0 / 9, 8.4 / 9}},
			{"password", "passw1ord", "", nil, e{1, 5, 3, 8.0 / 9, 8.4 / 9}},
			{"password", "passwo1rd", "", nil, e{1, 6, 2, 8.0 / 9, 8.4 / 9}},
			{"password", "passwor1d", "", nil, e{1, 7, 1, 8.0 / 9, 8.4 / 9}},
			{"password", "password1", "", nil, e{1, 8, 0, 8.0 / 9, 8.4 / 9}},
			{"password", "assword", "", nil, e{1, 0, 7, 7.0 / 8, 7.0 / 8}},
			{"password", "pssword", "", nil, e{1, 1, 6, 7.0 / 8, 7.1 / 8}},
			{"password", "pasword", "", nil, e{1, 3, 4, 7.0 / 8, 7.3 / 8}},
			{"password", "passord", "", nil, e{1, 4, 3, 7.0 / 8, 7.4 / 8}},
			{"password", "passwrd", "", nil, e{1, 5, 2, 7.0 / 8, 7.4 / 8}},
			{"password", "passwod", "", nil, e{1, 6, 1, 7.0 / 8, 7.4 / 8}},
			{"password", "passwor", "", nil, e{1, 7, 0, 7.0 / 8, 7.4 / 8}},

			// When a single character is replaced...
			{"password", "Xassword", "", nil, e{1, 0, 7, 7.0 / 8, 7.0 / 8}},
			{"password", "pXssword", "", nil, e{1, 1, 6, 7.0 / 8, 7.1 / 8}},
			{"password", "paXsword", "", nil, e{1, 2, 5, 7.0 / 8, 7.2 / 8}},
			{"password", "pasXword", "", nil, e{1, 3, 4, 7.0 / 8, 7.3 / 8}},
			{"password", "passXord", "", nil, e{1, 4, 3, 7.0 / 8, 7.4 / 8}},
			{"password", "passwXrd", "", nil, e{1, 5, 2, 7.0 / 8, 7.4 / 8}},
			{"password", "passwoXd", "", nil, e{1, 6, 1, 7.0 / 8, 7.4 / 8}},
			{"password", "passworX", "", nil, e{1, 7, 0, 7.0 / 8, 7.4 / 8}},

			// If characters are taken off the front and added to the back and all of
			// the characters are unique, then the distance is two times the number of
			// characters shifted, until you get halfway (and then it becomes easier
			// to shift from the other direction).
			{"12345678", "23456781", "", nil, e{2, 0, 0, 6. / 8, 6. / 8}},
			{"12345678", "34567812", "", nil, e{4, 0, 0, 4. / 8, 4. / 8}},
			{"12345678", "45678123", "", nil, e{6, 0, 0, 2. / 8, 2. / 8}},
			{"12345678", "56781234", "", nil, e{8, 0, 0, 0, 0}},
			{"12345678", "67812345", "", nil, e{6, 0, 0, 2. / 8, 2. / 8}},
			{"12345678", "78123456", "", nil, e{4, 0, 0, 4. / 8, 4. / 8}},
			{"12345678", "81234567", "", nil, e{2, 0, 0, 6. / 8, 6. / 8}},

			// If all the characters are unique and the values are reversed, then the
			// distance is the number of characters for an even number of characters,
			// and one less for an odd number of characters (since the middle
			// character will stay the same).
			{"12", "21", "", nil, e{2, 0, 0, 0, 0}},
			{"123", "321", "", nil, e{2, 0, 0, 1. / 3, 1. / 3}},
			{"1234", "4321", "", nil, e{4, 0, 0, 0, 0}},
			{"12345", "54321", "", nil, e{4, 0, 0, 1. / 5, 1. / 5}},
			{"123456", "654321", "", nil, e{6, 0, 0, 0, 0}},
			{"1234567", "7654321", "", nil, e{6, 0, 0, 1. / 7, 1. / 7}},
			{"12345678", "87654321", "", nil, e{8, 0, 0, 0, 0}},

			// The results are the same regardless of the string order,
			// with the default parameters...
			{"password", "1234", "", nil, e{8, 0, 0, 0, 0}},
			{"1234", "password", "", nil, e{8, 0, 0, 0, 0}},
			{"password", "pass1", "", nil, e{4, 4, 0, 4. / 8, 4. / 8}},
			{"pass1", "password", "", nil, e{4, 4, 0, 4. / 8, 4. / 8}},
			{"password", "passwor", "", nil, e{1, 7, 0, 7.0 / 8, 7.4 / 8}},
			{"passwor", "password", "", nil, e{1, 7, 0, 7.0 / 8, 7.4 / 8}},
			// ... but not necessarily so with custom costs:
			{"password", "1234", " (D=2)", NewParams().DelCost(2), e{12, 0, 0, 0, 0}},
			{"1234", "password", " (D=2)", NewParams().DelCost(2), e{8, 0, 0, 0, 0}},
			{"password", "pass1", " (D=2)", NewParams().DelCost(2), e{7, 4, 0, 4. / 11, 4. / 11}},
			{"pass1", "password", " (D=2)", NewParams().DelCost(2), e{4, 4, 0, 4. / 8, 4. / 8}},
			{"password", "pass1", " (S=3)", NewParams().SubCost(3), e{5, 4, 0, 8. / 13, 8. / 13}},
			{"password", "passwor", " (D=2)", NewParams().DelCost(2), e{2, 7, 0, 7.0 / 9, 7.8 / 9}},
			{"passwor", "password", " (D=2)", NewParams().DelCost(2), e{1, 7, 0, 7.0 / 8, 7.4 / 8}},

			// When setting a maxCost (should not affect Similarity() and Match())...
			{"password", "1password2", "(maxCost=6)", NewParams().MaxCost(6), e{2, 0, 0, 8. / 10, 8. / 10}},
			{"password", "pass1234", "(maxCost=1)", NewParams().MaxCost(1), e{2, 4, 0, 4. / 8, 4. / 8}},
			{"pass1word", "passwords1", "(maxCost=2)", NewParams().MaxCost(2), e{3, 4, 0, 7. / 10, 8.2 / 10}},
			{"password", "1passwo", " (D=2,maxCost=1)", NewParams().DelCost(2).MaxCost(1), e{2, 0, 0, 4. / 9, 4. / 9}},
			{"pwd", "password", " (I=0,maxCost=0)", NewParams().InsCost(0).MaxCost(0), e{0, 1, 1, 1, 1}},
			{"passXword", "password", "(maxCost=10)", NewParams().MaxCost(10), e{1, 4, 4, 8. / 9, 8.4 / 9}},
			{"passXord", "password", "(S=3,maxCost=17)", NewParams().SubCost(3).MaxCost(17), e{2, 4, 3, 14. / 16, 14.8 / 16}},
			// ... no change because the Calculate is calculated without getting into the main algorithm:
			{"password", "pass", "(maxCost=1)", NewParams().MaxCost(1), e{4, 4, 0, 4. / 8, 4. / 8}},
			{"password", "1234", " (D=2,maxCost=1)", NewParams().DelCost(2).MaxCost(1), e{8, 0, 0, 0, 0}},

			// When setting a minScore (should not affect Calculate() and Distance())...
			{"password", "pass1", "(minScore=0.3)", NewParams().MinScore(.3), e{4, 4, 0, 4. / 8, 4. / 8}},
			{"password", "pass1", "(minScore=0.6)", NewParams().MinScore(.6), e{4, 4, 0, 0, 0}},
			{"password", "pass1wor", "(minScore=0.9)", NewParams().MinScore(.9), e{2, 4, 0, 0, 0}},
			{"password", "password", "(minScore=1.1)", NewParams().MinScore(1.1), e{0, 8, 0, 0, 0}},

			// The rest of these are miscellaneous examples.  They will
			// be illustrated using the following key:
			// = (the characters are equal)
			// + (the character is inserted)
			// - (the character is removed)
			// # (the character is replaced)

			// Mississippi
			//  ippississiM
			// -=##====##=+ --> 6
			{"Mississippi", "ippississiM", "", nil, e{6, 0, 0, 5. / 11, 5. / 11}},

			// eieio
			// oieie
			// #===# --> 2
			{"eieio", "oieie", "", nil, e{2, 0, 0, 3. / 5, 3. / 5}},

			// brad+angelina
			// bra   ngelina
			// ===+++======= --> 3
			{"brad+angelina", "brangelina", "", nil, e{3, 3, 7, 10. / 13, 10.9 / 13}},

			// test international chars
			// naive
			// naïve
			// ==#== --> 1
			{"naive", "naïve", "", nil, e{1, 2, 2, 4. / 5, 4.2 / 5}},
		}
	)
	for _, c := range cases {
		par := c.p
		if par == nil {
			par = defaultParams
		}
		cost, lp, ls := Calculate([]rune(c.s1), []rune(c.s2), par.maxCost, par.insCost, par.subCost, par.delCost)
		if cost != c.exp.cost {
			t.Errorf("Cost: %q -> %q%s: got %d, want %d", c.s1, c.s2, c.desc, cost, c.exp.cost)
		}
		if lp != c.exp.lp {
			t.Errorf("Prefix: %q -> %q%s: got %d, want %d", c.s1, c.s2, c.desc, lp, c.exp.lp)
		}
		if ls != c.exp.ls {
			t.Errorf("Suffix: %q -> %q%s: got %d, want %d", c.s1, c.s2, c.desc, ls, c.exp.ls)
		}

		dist := Distance(c.s1, c.s2, c.p)
		if dist != c.exp.cost {
			t.Errorf("Distance: %q -> %q%s: got %d, want %d", c.s1, c.s2, c.desc, dist, c.exp.cost)
		}

		sim := Similarity(c.s1, c.s2, c.p)
		off := sim - c.exp.sim
		if off < 0 {
			off = -off
		}
		if off > 1e-15 {
			t.Errorf("Similarity: %q -> %q%s: got %f, want %f (off %g)", c.s1, c.s2, c.desc, sim, c.exp.sim, off)
		}

		match := Match(c.s1, c.s2, c.p)
		off = match - c.exp.match
		if off < 0 {
			off = -off
		}
		if off > 1e-15 {
			t.Errorf("Match: %q -> %q%s: got %f, want %f (off %g)", c.s1, c.s2, c.desc, match, c.exp.match, off)
		}
	}
}
