package portal

/*
ActorSignature is the generic type for structs that define the actions that are supported
by an actor. An actor signature can be used to invoke remote actor actions in a type-safe way.

Darlean uses reflection to derive the actor type and the names and argument and result types of
action methods.

Therefore, actor signatures must obey the following rules:

* An actor signature must be a a struct
* The struct must be named `<ActorType>`. For example, `FriendlyActor`. Note that actor types in
  Darlean are case-insensitive, and that only letters and digits are preserved.
* The struct must contain one field of type [portal.ActionSignature] for each actor action. The name
  of the field must match with the name of the actor action. Action names are case-insensitive and
  other characters than letters and digits are omitted.

Example:

	type FriendlyActor struct {
		Echo  EchoActor_Echo
		Greet EchoActor_Greet
	}
*/
type ActorSignature any

/*
ActionSignature is the generic type for a struct that defines the attributes and return type
for one specific acor action. They are intended to be contained within a [portal.ActorSignature].

Darlean uses reflection to derive the actor type, action name, argument types and result type of
the action.

Therefore, action signatures must obey the following rules:

* An actor signature must be a struct named `<ActorType>_<ActionName>`. For example, `FriendlyActor_Greet`.
* The struct must have one field for each attribute. The field must be named `A0` for the first call attribute,
  `A1` for the second attribute, and so on.
* For convenience, the name of the attribute can be provided for after an underscore: `A0_Foo`.
* The field must be of the proper type for the attribute. Supported primitive types are strings, numbers, booleans,
  [github.com/darlean-io/darlean.go/utils/binary/Binary]. Supported compound types are structs, maps and slices of supported types.
* When the action is not a void, the struct should define a field called `Result` field of the proper type.
*/
type ActionSignature any
