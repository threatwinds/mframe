package mframe

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quantfall/rerror"
)

func (d *DataFrame) Filter(operator, key string, value interface{}, options map[string]bool) *DataFrame {
	d.Locker.RLock()
	defer d.Locker.RUnlock()

	var keys = make(map[string]string)

	if Contains(key, "^") || Contains(key, "[") || Contains(key, "(") {
		for k, t := range d.Keys {
			if MatchesRegExp(k, key) {
				keys[k] = t
			}
		}
	} else {
		keys[key] = d.Keys[key]
	}

	var results = new(DataFrame)
	results.Init(10 * time.Minute)

	for k, t := range keys {
		switch t {
		case "numeric":
			aValue, ok := value.(float64)
			if !ok {
				return results
			}
			switch operator {
			case "==":
				if ids, ok := d.Numerics[k][aValue]; ok {
					for id := range ids {
						results.Insert(d.Data[id])
					}
				}
			case "!=":
				if v, ok := d.Numerics[k]; ok {
					for v, ids := range v {
						if Equals(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case ">":
				if v, ok := d.Numerics[k]; ok {
					for v, ids := range v {
						if !MajorThan(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "<":
				if v, ok := d.Numerics[k]; ok {
					for v, ids := range v {
						if MajorThan(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case ">=":
				if v, ok := d.Numerics[k]; ok {
					for v, ids := range v {
						if !Equals(v, aValue) && !MajorThan(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "<=":
				if v, ok := d.Numerics[k]; ok {
					for v, ids := range v {
						if !Equals(v, aValue) && MajorThan(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "in list":
				aValue, ok := value.([]float64)
				if !ok {
					return results
				}
				if v, ok := d.Numerics[k]; ok {
					for v, ids := range v {
						if !InList(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "not in list":
				aValue, ok := value.([]float64)
				if !ok {
					return results
				}
				if v, ok := d.Numerics[k]; ok {
					for v, ids := range v {
						if InList(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			default:
				rerror.ErrorF(http.StatusBadRequest, "incorrect operator '%s' for key '%s' of type '%s'", operator, key, t)
			}
		case "string":
			aValue, ok := value.(string)
			if !ok {
				return results
			}
			switch operator {
			case "==":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							aValue = strings.ToLower(aValue)
						}

						if !Equals(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "!=":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							aValue = strings.ToLower(aValue)
						}

						if Equals(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "regexp":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if !MatchesRegExp(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "not regexp":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if MatchesRegExp(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "in list":
				aValue, ok := value.([]string)
				if !ok {
					return results
				}
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							for k, v := range aValue {
								aValue[k] = strings.ToLower(v)
							}
						}

						if !InList(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "not in list":
				aValue, ok := value.([]string)
				if !ok {
					return results
				}
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							for k, v := range aValue {
								aValue[k] = strings.ToLower(v)
							}
						}

						if InList(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "in cidr":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if !InCIDR(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "not in cidr":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if InCIDR(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "contains":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							value = strings.ToLower(aValue)
						}

						if !Contains(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "not contains":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							value = strings.ToLower(aValue)
						}

						if Contains(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "starts with":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							value = strings.ToLower(aValue)
						}

						if !StartsWith(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "not starts with":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							value = strings.ToLower(aValue)
						}

						if StartsWith(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "ends with":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							value = strings.ToLower(aValue)
						}

						if !EndsWith(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			case "not ends with":
				if v, ok := d.Strings[k]; ok {
					for v, ids := range v {
						if sensitive, ok := options["case-sensitive"]; ok && !sensitive {
							v = strings.ToLower(v)
							value = strings.ToLower(aValue)
						}

						if EndsWith(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			default:
				rerror.ErrorF(http.StatusBadRequest, "incorrect operator '%s' for key '%s' of type '%s'", operator, key, t)
			}
		case "boolean":
			aValue, ok := value.(bool)
			if !ok {
				return results
			}
			switch operator {
			case "==":
				if ids, ok := d.Booleans[k][aValue]; ok {
					for id := range ids {
						results.Insert(d.Data[id])
					}
				}
			case "!=":
				if v, ok := d.Booleans[k]; ok {
					for v, ids := range v {
						if Equals(v, aValue) {
							continue
						}

						for id := range ids {
							results.Insert(d.Data[id])
						}
					}
				}
			default:
				rerror.ErrorF(http.StatusBadRequest, "incorrect operator '%s' for key '%s' of type '%s'", operator, key, t)
			}
		}
	}

	return results
}

func (d *DataFrame) FindFirstByKey(key string) (uuid.UUID, string, interface{}) {
	d.Locker.RLock()
	defer d.Locker.RUnlock()

	var keys = make(map[string]string)

	if Contains(key, "^") || Contains(key, "[") || Contains(key, "(") {
		for k, t := range d.Keys {
			if MatchesRegExp(k, key) {
				keys[k] = t
			}
		}
	} else {
		keys[key] = d.Keys[key]
	}

	for k, t := range keys {
		switch t {
		case "numeric":
			if values, ok := d.Numerics[k]; ok {
				for _, value := range values {
					for row := range value {
						return row, k, d.Data[row][k]
					}
				}
			}
		case "string":
			if values, ok := d.Strings[k]; ok {
				for _, value := range values {
					for row := range value {
						return row, k, d.Data[row][k]
					}
				}
			}
		case "boolean":
			if values, ok := d.Booleans[k]; ok {
				for _, value := range values {
					for row := range value {
						return row, k, d.Data[row][k]
					}
				}
			}
		}
	}

	return uuid.Nil, key, new(DataFrame)
}

func Equals(left, right interface{}) bool {
	return left == right
}

func MatchesRegExp(value, regExp string) bool {
	re, err := regexp.Compile(regExp)
	if err == nil {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func MajorThan(left, right float64) bool {
	return left > right
}

func InList[v float64 | string, result bool](value v, list []v) result {
	for _, element := range list {
		if element == value {
			return true
		}
	}
	return false
}

func InCIDR(value, cidr string) bool {
	_, subnet, err := net.ParseCIDR(cidr)
	if err == nil {
		ip := net.ParseIP(value)
		if ip != nil {
			if subnet.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func Contains(value, substring string) bool {
	return strings.Contains(value, substring)
}

func StartsWith(value, prefix string) bool {
	return strings.HasPrefix(value, prefix)
}

func EndsWith(value, suffix string) bool {
	return strings.HasSuffix(value, suffix)
}
