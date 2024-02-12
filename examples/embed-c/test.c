#include "embedlib.h"
#include <stdio.h>
#include <windows.h>
#include <stdint.h>


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

handle app = 0;

// Handler when a remote invoke operation is done
void onInvoked(handle call, GoString response) {
    char *resp = cstr(response);

    printf("[%i] Invoked: %s\n", call, resp);
}

// Handle an incoming action. We should forward the call to one of our internal actors/actions/methods and
// when finished, invoke the SubmitActionResult.
void handleAction(handle action, handle call, GoString request) {
    char *req = cstr(request);
    
    printf("[%i] We should be handling action: %s. For now, return a dummy Bar result\n", call, req);

    GoString* result = gostr("{\"Value\": \"Bar\"}");
    SubmitActionResult(call, *result);
}

int main(void) {

    // Create the app with some basic config
    app = CreateApp(*gostr(appId), *gostr(nats), *gostr(nodes));

    // Register our local actor.
    // Note: This functionality is not yet implemented in the go api wrapper.
    GoString* ourActorRegistration = gostr("{\"actorType\": \"ouractor\"}");
    GoString* ourActionRegistation = gostr("{\"actionName\": \"echo\", \"locking\": \"exclusive\"}");
    handle actor = RegisterActor(app, *ourActorRegistration);
    handle action = RegisterAction(actor, *ourActionRegistation, handleAction);

    // Start the app.
    StartApp(app);

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
    Invoke(app, 1, onInvoked, *gostr("{\"actorType\": \"typescriptactor\", \"actorId\": [\"a\"], \"actionName\": \"echo\", \"arguments\": [\"Moon\"]}")); 
    Invoke(app, 2, onInvoked, *gostr("{\"actorType\": \"ouractor\", \"actorId\": [\"a\"], \"actionName\": \"echo\", \"arguments\": [\"Moon\"]}")); 
    
    Sleep(5*1000);
    
    StopApp(app);
    ReleaseApp(app);
}