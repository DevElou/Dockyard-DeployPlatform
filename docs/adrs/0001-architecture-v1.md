# ADR 0001: Architecture V1 de Dockyard

## Statut

Accepted

## Date

2026-04-15

## Contexte

Dockyard doit fournir une plateforme privee de deploiement pour un usage restreint a deux operateurs. La cible court terme est une V1 exploitable sur infrastructure Docker existante. La cible moyen terme est de preparer une evolution vers Kubernetes sans reecrire le coeur metier.

Le systeme doit :

- connecter un repository GitHub
- construire une image versionnee
- pousser cette image dans un registry prive
- deployer sur un ou plusieurs serveurs Docker
- exposer l'application via un domaine
- permettre rollback et suivi d'etat

## Decision

Dockyard V1 adopte une architecture separee en quatre blocs :

1. `control plane`
2. `orchestration / workers`
3. `deployment agents`
4. `edge / persistence`

Choix techniques :

- frontend : Next.js
- API : Go
- worker : Go separe
- base principale : CockroachDB
- queue : Redis + asynq
- runtime V1 : Docker
- edge : Nginx Proxy Manager
- registry : Docker Registry prive
- source provider V1 : GitHub
- DNS provider V1 : DuckDNS

## Raisons

### Pourquoi Go pour API, worker et agent

- uniformite technique entre API, worker et agent
- meilleur controle de la concurrence et des binaires long-lived
- binaire leger et simple a deployer sur les hosts
- frontieres plus nettes entre domaine et adapters si la structure est tenue

### Pourquoi separer API et worker

- l'API reste synchrone et lisible
- les workflows longs sont isoles
- les pannes de build/deploiement ne degradent pas directement l'API

### Pourquoi un agent par host

- pas d'acces SSH a disperser dans le control plane
- meilleure isolation des responsabilites
- facilite la transition vers d'autres runtimes
- implementation naturelle en Go avec un binaire simple

### Pourquoi Nginx Proxy Manager

- deja adapte a une infrastructure personnelle ou homelab
- interface d'administration utile pendant la V1
- API exploitable par Dockyard pour creer et mettre a jour les proxy hosts
- gestion TLS integree sans imposer un routage par labels Docker

### Pourquoi CockroachDB

- apprentissage souhaite
- SQL relationnel
- trajectoire naturelle vers une persistence distribuee

## Consequences

### Positives

- architecture modulaire mais pas surconcue
- frontieres claires entre metier et execution infra
- bonne trajectoire vers un `RuntimeDriver` Kubernetes

### Negatives

- complexite superieure a un simple script Compose
- ajout d'un agent a maintenir
- discipline necessaire sur les transactions CockroachDB

## Alternatives ecartees

### 1. Monolithe unique avec API + jobs inline

Refuse car :
- melange des responsabilites
- plus difficile a stabiliser
- moins bon controle de l'execution des workflows

### 2. SSH depuis le backend vers les hosts

Refuse car :
- fragilise la securite
- cree un couplage direct backend -> infra
- complique la future abstraction runtime

### 3. Docker Compose comme primitive metier principale

Refuse car :
- l'abstraction cible est `Release -> Deployment`, pas `compose up`
- Compose peut etre un detail d'implementation local, pas le modele central

### 4. Kubernetes des la V1

Refuse car :
- trop de cout de complexite pour le besoin immediat
- reduit la vitesse de livraison sans benefice produit court terme

## Regles structurantes

- `Release` est immutable
- `Deployment` est un evenement d'application d'une release sur une cible
- `Domain` est une ressource metier explicite
- le worker orchestre
- l'agent execute
- le runtime concret est derriere une abstraction
