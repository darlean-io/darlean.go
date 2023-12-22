package remoteactorregistry

const SERVICE = "io.darlean.actorregistryservice"

type ObtainRequest struct {
}

type ApplicationInfo struct {
	Name             string  `json:"name"`
	MigrationVersion *string `json:"migrationVersion"`
}

type ActorPlacement struct {
	AppBindIdx *int  `json:"appBindIdx"`
	Sticky     *bool `json:"sticky"`
}

type ActorInfo struct {
	Applications []ApplicationInfo `json:"applications"`
	Placement    ActorPlacement    `json:"Placement"`
}

type ObtainResponse struct {
	Nonce     string               `json:"nonce"`
	ActorInfo map[string]ActorInfo `json:"actorInfo"`
}

type ActorPushInfo struct {
	Placement        ActorPlacement `json:"placement"`
	MigrationVersion string         `json:"migrationVersion"`
}

type PushRequest struct {
	Application string                   `json:"application"`
	ActorInfo   map[string]ActorPushInfo `json:"actorInfo"`
}

const ACTION_OBTAIN = "obtain"
const ACTION_PUSH = "push"
