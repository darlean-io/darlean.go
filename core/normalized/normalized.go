package normalized

import "strings"

type ActorType string
type ActionName string

func NormalizeActorType(name string) ActorType {
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, ".", "")
	result = strings.TrimFunc(result, func(r rune) bool {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		return !valid
	})
	return ActorType(result)
}

func NormalizeActionName(name string) ActionName {
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, ".", "")
	result = strings.TrimFunc(result, func(r rune) bool {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		return !valid
	})
	return ActionName(result)
}
