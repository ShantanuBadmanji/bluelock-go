package set

// Set is a custom type to represent a set of comparable types
type Set[T comparable] map[T]struct{}

// New creates a new empty set
func New[T comparable]() Set[T] {
	return Set[T]{}
}

// NewWithCapacity creates a new set with a specified initial capacity
// Note: The capacity is not a strict limit, as Go maps can grow dynamically.
// This is just a hint to optimize memory allocation.
func NewWithCapacity[T comparable](capacity int) Set[T] {
	return make(Set[T], capacity)
}

// NewFromSlice creates a new set from a slice of elements
// It adds all elements from the slice to the set.
// Note: Duplicates in the slice will be ignored.
func NewFromSlice[T comparable](slice []T) Set[T] {
	set := New[T]()
	set.AddAll(slice...)
	return set
}

// Add adds an element to the set
func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

// AddAll adds multiple elements to the set
func (s Set[T]) AddAll(values ...T) {
	for _, value := range values {
		s.Add(value)
	}
}

// Remove deletes an element from the set
func (s Set[T]) Remove(value T) {
	delete(s, value)
}

// Contains checks if an element exists in the set
func (s Set[T]) Contains(value T) bool {
	_, exists := s[value]
	return exists
}

// Size returns the number of elements in the set
func (s Set[T]) Size() int {
	return len(s)
}

// Clear removes all elements from the set
func (s Set[T]) Clear() {
	for key := range s {
		delete(s, key)
	}
}

// IsEmpty checks if the set is empty
func (s Set[T]) IsEmpty() bool {
	return len(s) == 0
}

// Union returns a new set that is the union of two sets
func (s Set[T]) Union(other Set[T]) Set[T] {
	unionSet := New[T]()
	for key := range s {
		unionSet.Add(key)
	}
	for key := range other {
		unionSet.Add(key)
	}
	return unionSet
}

// Intersection returns a new set that is the intersection of two sets
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	intersectionSet := New[T]()
	for key := range s {
		if other.Contains(key) {
			intersectionSet.Add(key)
		}
	}
	return intersectionSet
}

// Difference returns a new set that is the difference of two sets
func (s Set[T]) Difference(other Set[T]) Set[T] {
	differenceSet := New[T]()
	for key := range s {
		if !other.Contains(key) {
			differenceSet.Add(key)
		}
	}
	return differenceSet
}
