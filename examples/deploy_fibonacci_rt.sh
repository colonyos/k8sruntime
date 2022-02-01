#!/bin/bash

export COLONYID="bce704f7995b6b2af337621fe48d0f28da25bf8de6d1270ddf79ed4354f32645"
export RUNTIMEID="e91c4506cbdd01900bd65180dd660d92c2b6c5fd406741da65acdc74e7dd906b"

colonies process submit --spec kube_process_spec.json 
