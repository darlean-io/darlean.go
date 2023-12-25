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

type ActorPushInfo struct {
	Placement        ActorPlacement
	MigrationVersion string
}

type ActorRegistryFetcher interface {
	Get(actorType string) *ActorInfo
}

type ActorRegistryPusher interface {
	Set(info map[string]ActorPushInfo)
}
