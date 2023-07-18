#!/usr/bin/env bash

COUNT=$(($1 + 1))

TASKID=$(rally task list | tail -$COUNT | head -1 | awk '{print $2}')

rally task delete --uuid $TASKID
