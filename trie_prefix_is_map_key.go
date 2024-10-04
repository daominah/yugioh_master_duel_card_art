package yugioh_master_duel_card_art

type TrieNode struct {
	children    map[rune]*TrieNode
	isEndOfWord bool
}

// Trie represents the Trie itself
type Trie struct {
	root *TrieNode
}

// NewTrie initializes a new Trie
func NewTrie() *Trie {
	return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

// Insert adds a word to the Trie
func (t *Trie) Insert(word string) {
	node := t.root
	for _, char := range word {
		if _, exists := node.children[char]; !exists {
			node.children[char] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[char]
	}
	node.isEndOfWord = true
}

// CheckPrefixIsAKey checks if any prefix of the string exists in the Trie
func (t *Trie) CheckPrefixIsAKey(s string) bool {
	node := t.root
	for _, char := range s {
		if _, exists := node.children[char]; !exists {
			return false
		}
		node = node.children[char]
		if node.isEndOfWord {
			return true
		}
	}
	return false
}

// ContainsMapKeyTrie checks if the string starts with any key from the map using a Trie
func ContainsMapKeyTrie(s string, myMap map[string]int) bool {
	trie := NewTrie()
	for key := range myMap {
		trie.Insert(key)
	}
	return trie.CheckPrefixIsAKey(s)
}
