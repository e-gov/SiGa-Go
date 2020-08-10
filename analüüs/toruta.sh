#!/bin/bash

OR='\033[0;07m' # Reversed
NC='\033[0m' # (No Color)

echo -e "${OR} GNU utiliitide abil uurin koodi"
echo -e " {NC}"

grep "func " -r siga/*.go
