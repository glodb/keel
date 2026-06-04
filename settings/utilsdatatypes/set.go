package utilsdatatypes

import "github.com/bytedance/sonic"

type Set[T comparable] struct {
	data map[T]bool
}

// NewSet creates a new Set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		data: make(map[T]bool),
	}
}

// Add adds an element to the Set
func (set Set[T]) Add(element T) {
	set.data[element] = true
}

// AddSlice adds multiple elements from a slice to the Set
func (set *Set[T]) AddSlice(elements []T) {
	for _, element := range elements {
		set.Add(element)
	}
}

// Contains checks if an element is present in the Set
func (set *Set[T]) Contains(element T) bool {
	return set.data[element]
}

// Remove removes an element from the Set
func (set *Set[T]) Remove(element T) {
	delete(set.data, element)
}

// Clear removes all elements from the Set
func (set *Set[T]) Clear() {
	set.data = make(map[T]bool)
}

// Size returns the size of the Set
func (set *Set[T]) Size() int {
	return len(set.data)
}

func (set *Set[T]) GetMap() map[T]bool {
	return set.data
}

func (s *Set[T]) MarshalJSON() ([]byte, error) {
	// Convert map to slice of strings
	var keys []T
	for key := range s.data {
		keys = append(keys, key)
	}

	return sonic.Marshal(keys)
}

func (set *Set[T]) UnmarshalJSON(data []byte) error {
	// Initialize the map if it's nil
	if set.data == nil {
		set.data = make(map[T]bool)
	}

	if string(data) == "null" || string(data) == "[]" || string(data) == "{}" {
		return nil // Set remains empty
	}
	// Decode JSON array into a slice of strings
	var keys []T
	if err := sonic.Unmarshal(data, &keys); err != nil {
		return err
	}

	// Populate the set
	for _, key := range keys {
		set.data[key] = true
	}

	return nil
}

// List returns the elements of the Set as a slice of strings
func (set *Set[T]) List() []T {
	keys := make([]T, 0, len(set.data))
	for key := range set.data {
		keys = append(keys, key)
	}
	return keys
}
