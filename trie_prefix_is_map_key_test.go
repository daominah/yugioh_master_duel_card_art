package yugioh_master_duel_card_art

import (
	"testing"
)

func TestTrie(t *testing.T) {
	validNames := []string{
		"beatrice__lady_of_the_eternal",
		"blue_eyes_white_dragon",
		"super_starslayer_ty_phon___sky_crisis",
	}
	trie := NewTrie()
	for _, name := range validNames {
		trie.Insert(name)
	}

	for _, c := range []struct {
		s    string
		want bool
	}{
		{"blue_eyes_white_dragon_4007_a1.png", true},
		{"blue_eyes_white_dragon_a1_transparent_blender.png", true},
		{"blue_eyes", false},
		{"super_starslayer_ty_phon___sky_crisis_19184_up2048.png", true},
		{"beatrice__lady_of_the_eternal_12108_ocg_up2048.png", true},
		{"burning_abyss_beatrice__lady_of_the_eternal_12108_ocg_up2048.png", false},
	} {
		got := trie.CheckPrefixIsAKey(c.s)
		if got != c.want {
			t.Errorf("error CheckPrefixIsAKey(%v) got %v, but want %v", c.s, got, c.want)
		}
	}
}
