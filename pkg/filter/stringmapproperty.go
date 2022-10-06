package filter

import (
	"fmt"
	"github.com/opencost/opencost/pkg/kubecost"
	"github.com/opencost/opencost/pkg/log"
	"strings"
)

type StringMapPropertied interface {
	StringMapProperty(string) (map[string]string, error)
}

// StringMapOperation is an enum that represents operations that can be performed
// when filtering (equality, inequality, etc.)
type StringMapOperation string

const (
	// StringMapHasKey passes if the map has the provided key
	StringMapHasKey StringMapOperation = "stringmapcontains"

	StringMapStartsWith = "stringmapstartswith"

	// StringMapEquals when the given key and value match
	StringMapEquals = "stringmapequals"
)

// StringMapProperty is the lowest-level type of filter. It represents
// a filter operation (equality, inequality, etc.) on a property that contains a string slice
type StringMapProperty[T StringMapPropertied] struct {
	Field string
	Op    StringMapOperation
	Key   string
	Value string
}

func (smp StringMapProperty[T]) String() string {
	return fmt.Sprintf(`(%s %s[%s] "%s")`, smp.Op, smp.Field, smp.Key, smp.Value)
}

// Flattened returns itself because you cannot flatten a base condition further
func (smp StringMapProperty[T]) Flattened() Filter[T] {
	return smp
}

func (smp StringMapProperty[T]) equals(that Filter[T]) bool {
	if thatAC, ok := that.(StringMapProperty[T]); ok {
		return smp == thatAC
	}

	return false
}

func (smp StringMapProperty[T]) Matches(that T) bool {

	thatMap, err := that.StringMapProperty(smp.Field)
	if err != nil {
		log.Errorf("Filter: StringMapProperty: could not retrieve field %s: %s", smp.Field, err.Error())
		return false
	}

	valueToCompare, keyIsPresent := thatMap[smp.Key]

	switch smp.Op {
	case StringMapHasKey:
		return keyIsPresent
	case StringMapEquals:
		// namespace:"__unallocated__" should match a.Properties.Namespace = ""
		// label[app]:"__unallocated__" should match _, ok := Labels[app]; !ok
		if !keyIsPresent || valueToCompare == "" {
			return smp.Value == kubecost.UnallocatedSuffix
		}

		if valueToCompare == smp.Value {
			return true
		}

	case StringMapStartsWith:
		if !keyIsPresent {
			return false
		}

		// We don't need special __unallocated__ logic here because a query
		// asking for "__unallocated__" won't have a wildcard and unallocated
		// properties are the empty string.

		return strings.HasPrefix(valueToCompare, smp.Value)
	default:
		log.Errorf("Filter: StringMapProperty: Unhandled filter op. This is a filter implementation error and requires immediate patching. Op: %s", smp.Op)
		return false
	}

	return false
}