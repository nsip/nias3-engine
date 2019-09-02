# nias3-engine

*** DEPRECATED (RETIRED) ***

*This repository is obsolete and retired (archived). This is an unmantained repository. In particular, note that it is not being updated with security patches, including those for dependent libraries.*


p2p and sig-chain core engine for n3 services

Presupposes the following are installed:

* NATS Streaming (gets installed by build.sh and compiled)
* rethinkdb

05-2018
Check-in of code to date so that forward changes can be properly branched. Not to be considererd a release candidate yet, but most of the key pieces are in place.

The NIAS3 Engine needs to process a large number of open files, since it fields a large number of API requests through its webserver. If you will be running NIAS3 on Mac/Linux, you will need to increase your ulimit setting; we recommend 2048.

* For Mac, see https://blog.dekstroza.io/ulimit-shenanigans-on-osx-el-capitan/
* For Linux, `ulimit -n 2048`

