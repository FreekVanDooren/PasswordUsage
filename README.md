# Password Usage Checker
Based on [this video](https://www.youtube.com/watch?v=hhUb5iknVJs) and [this tutorial](https://medium.com/google-cloud/kubernetes-110-your-first-deployment-bf123c1d3f8) I thought I'd experiment a bit with Kubernetes (locally) and see how far I'll get with Go this time.

## Summary
Application checks if a password has been 'pwnd' either on command line or via HTTP endpoint. 

##Tools used
| Tool | Version |
| ---- | ------- |
| Go | go1.12.1 |
| Minikube | v1.0.0 |
| Docker | 18.09.0 |
| Kubectl  (client) | v1.11.7-dispatcher |
| Kubectl  (server) | v1.14.0 |

## How to run
```
$ docker-compose up -d --build
```
### For cli
```
$ docker attach password-checker
```
_NB_
1. Console will be empty at first, but typing password to be checked will be responsive.
1. Exit attached state with `ctrl+p` `ctrl+q`
### For web
```
$ curl localhost:6544/password-checker/<password>
```
Where `<password>` is to be replaced with desired.
## To Do
- Get to the end of the tutorial
- Split up cli from core functionality (Introduce types)