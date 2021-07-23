#!/bin/bash

declare -a pids
pids=`ps aguwx | grep run_polls.sh | awk '{print $2}' | xargs`
for thisPid in "${pids[@]:0}"
do
	printf "killing $thisPid \n"
	kill -9 $thisPid
done;