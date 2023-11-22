package actorregistry

type ApplicationInfo struct {
	Name             string
	MigrationVersion *string
}

type ActorPlacement struct {
	AppBindIdx *int
	Sticky     *bool
}

type ActorInfo struct {
	Applications []ApplicationInfo
	Placement    ActorPlacement
}

type ActorRegistry interface {
	Get(actorType string) *ActorInfo
}
