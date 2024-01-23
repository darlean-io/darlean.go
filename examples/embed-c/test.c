#include "embedlib.h"
#include <stdio.h>
#include<windows.h>

// See: https://github.com/vladimirvivien/go-cshared-examples/blob/master/README.md

char appId[] = "client";
char nats[] = "localhost:4500";
char nodes[] = "server";

// Convert c string into a go string.
GoString* gostr(char* content) {
    GoString* gs = malloc(sizeof(GoString));
    gs->p = content;
    gs->n = strlen(content);
    return gs;
}

// Convert go string into a c string.
char* cstr(GoString s) {
    char *value = malloc(s.n+1);
    memcpy(value, s.p, s.n);
    value[s.n] = '\0';
    return value;
}


// Handler when a remote invoke operation is done
void onInvoked(GoString s) {
    char *s2 = cstr(s);

    printf("Invoked: %s\n", s2);
}

// Handle an incoming action. We should forward the call to one of our internal actors/actions/methods and
// when finished, invoke the SubmitActionResult.
void handleAction(GoString s) {
    char *s2 = cstr(s);

    printf("We should be handling action: %s\n", s2);

    GoString* callId = gostr("1234567");
    GoString* result = gostr("{\"Result\": \"Bar\"}");
    SubmitActionResult(*gostr(appId), *callId, *result);
}

int main(void) {

    // Create the app with some basic config
    CreateApp(*gostr(appId), *gostr(nats), *gostr(nodes));

    // Register our local actor.
    // Note: This functionality is not yet implemented in the go api wrapper.
    GoString* ourActorRegistration = gostr("{\"actorType\": \"ouractor\"}");
    GoString* ourActionRegistation = gostr("{\"actorType\": \"ouractor\", \"actionName\": \"echo\", \"locking\": \"exclusive\"}");
    RegisterActor(*gostr(appId), *ourActorRegistration);
    RegisterAction(*gostr(appId), *ourActionRegistation, handleAction);

    // Start the app.
    StartApp(*gostr(appId));

    // Invoke a remote actor
    GoString* goActorType = gostr("typescriptactor");
    GoString* goActorId0 = gostr("a");
    GoString* goActionName = gostr("echo");
    GoString* goArgument = gostr("Moon");
    
    GoString* goActorIdList[1] = {goActorId0};
    GoSlice goActorId = {data: *goActorIdList, 1, 1};
    GoString* goArgumentList[1] = {goArgument};
    GoSlice goArguments = {data: *goArgumentList, 1, 1};


    printf("Invoking: %s\n", cstr(*goArgument));
    Invoke(*gostr(appId), onInvoked,  *goActorType, goActorId, *goActionName, *goArgument);
    
    Sleep(5*1000);
    
    StopApp(*gostr(appId));
    ReleaseApp(*gostr(appId));
}