# Introduction

[Darlean](https://darlean.io) is an open source, cross-language library for creating distributed backend applications. It eliminates much of the
complexity that comes with distributed scalable computing (like inter process comunnication, scalability, concurrency, deployment and persistence) so that
engineers can focus on the actual domain logic.

Darlean is created to solve the [microservice premium](https://martinfowler.com/bliki/MicroservicePremium.html). 
Instead of having to start a project as a complex (and expensive) 
microservice project on forehand because it may grow in the future, or to start a project as a simple monolith at the risk of having to perform an expensive
rewrite when scalability or availability actually become an issue, Darlean makes it possible to start simple and to scale out when necessary with
zero or minimal changes to your application.

The library provides:
* [Virtual actor](https://darlean.io/the-virtual-actor-model/) primitives that are well integrated with the supported programming languages (currently TS/JS and Go).
* An [integrated high-performance message bus](https://darlean.io/documentation/configuration-options/#messaging-options) for inter-process communication (an external NATS message bus can also be configured).
* Integrated [scalable persistence](https://darlean.io/documentation/persistence/) (extendable architecture; external persistence providers can be used as well)
* Integrated [scalable indexed tables](https://darlean.io/documentation/tables/) (extendable architecture; external table services can be used as well)
* Integrated api gateways allow invocation of actors via HTTP/S.
* Persistent timers that invoke actors even when they are asleep.

# Getting Started
TODO: Guide users through getting your code up and running on their own system. In this section you can talk about:
1.	Installation process
2.	Software dependencies
3.	Latest releases
4.	API references

# Build and Test
TODO: Describe and show how to build your code and run the tests. 

# Contribute
TODO: Explain how other users and developers can contribute to make your code better. 

If you want to learn more about creating good readme files then refer the following [guidelines](https://docs.microsoft.com/en-us/azure/devops/repos/git/create-a-readme?view=azure-devops). You can also seek inspiration from the below readme files:
- [ASP.NET Core](https://github.com/aspnet/Home)
- [Visual Studio Code](https://github.com/Microsoft/vscode)
- [Chakra Core](https://github.com/Microsoft/ChakraCore)