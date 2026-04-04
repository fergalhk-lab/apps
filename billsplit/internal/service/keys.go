// billsplit/internal/service/keys.go
package service

func groupKey(id string) string { return "groups/" + id + ".json" }

func removeString(ss []string, s string) []string {
	out := make([]string, 0, len(ss))
	for _, v := range ss {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
