#include "embedlib.h"
#include <stdio.h>
#include<windows.h>

// See: https://github.com/vladimirvivien/go-cshared-examples/blob/master/README.md

void onInvoked(GoString s) {
    char value[s.n+1];
    memcpy(value, s.p, s.n);
    value[s.n] = '\0';

    printf("Invoked: %s\n", value);
}

int main(void) {
    char appId[] = "client";
    char nats[] = "localhost:4500";
    char nodes[] = "server01";

    GoString goAppId = {p: appId, n: sizeof(appId)-1};
    GoString goNats = {p: nats, n: sizeof(nats)-1};
    GoString goNodes = {p: nodes, n: sizeof(nodes)-1};
    
    Start(goAppId, goNats, goNodes);

    char actorType[] = "echoactor";
    char actorId0[] = "a";
    char actionName[] = "echo";
    char argument[] = "Moon";


    GoString goActorType = {p: actorType, n: sizeof(actorType)-1};

    GoString goActorId0 = {p: actorId0, n: sizeof(actorId0)-1};
    GoString goActorIdList[1] = {goActorId0};
    GoSlice goActorId = {data: goActorIdList, 1, 1};

    GoString goActionName = {p: actionName, n: sizeof(actionName)-1};

    GoString goArgument = {p: argument, n: sizeof(argument)-1};
    GoString goArgumentList[1] = {goArgument};
    GoSlice goArguments = {data: goArgumentList, 1, 1};

    Invoke(onInvoked, goActorType, goActorId, goActionName, goArguments);
    
    Sleep(5*1000);
    
    Stop();
}