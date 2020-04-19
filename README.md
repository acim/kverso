# kverso

Kubernetes image versions manager

[![HitCount](http://hits.dwyl.com/acim/kverso.svg)](http://hits.dwyl.com/acim/kverso)

## Installation

kubectl apply -f https://raw.githubusercontent.com/acim/kverso/master/deploy/rbac.yaml

kubectl apply -f https://raw.githubusercontent.com/acim/kverso/master/deploy/deployment.yaml

## Usage

Find pod name using kubectl get pod

kubectl port-forward pod/kverso-name 3000:3000

Point your browser to localhost:3000

## Think about using

https://github.com/docker/distribution/blob/master/registry/client/repository.go
https://github.com/mcuadros/go-version
