#!/bin/bash

OR='\033[0;07m' # Reversed
NC='\033[0m' # (No Color)

echo -e "${OR} Moodustan paki siga väljakutsegraafi"
echo -e " Eeldus: go tools vahend cmd/callgraph paigaldatud."
echo -e " Moodustan failid: analyys/graph.txt (vahefail) ja"
echo -e " analyys/pure.txt (lõppfail). Lõppfaili alusel saab genereerida GraphViz"
echo -e " diagrammi (online-teenus https://dreampuf.github.io/GraphvizOnline/).{NC}"
echo -e " Diagramm siiski ei ole väga informatiivne.{NC}"

callgraph \
  -format graphviz \
  -algo static \
  main.go apphelpers.go signwithidcard.go signwithmid.go | \
  grep -e "siga.*siga" -e "digraph" -e "}" \
  > analyys/graph.txt


# Selgitus:
# 1) väljasta GraphViz formaadis
# 2) algo "static"
# 3) tuleb anda kõik paki main failid
# 4) vali väljakutsed siga -> siga, säilitades ka GraphViz sisendfaili päise
# (digraph) ja lõpu (loogeline sulg).

echo -e "${OR} Teisendan cmd/callgraph väljundi (fail graph.txt)"
echo -e " faili pure.txt ${NC}"

cat analyys/graph.txt | \
  awk '!/https/' | \
  awk '!/confutil/' | \
  sed  's/github.com\/e-gov\/SiGa-Go\///g' | \
  sed  's/(\*//g' | sed  's/)//g' \
  > analyys/pure.txt

echo -e "${OR} Lõpp ${NC}"

# Selgitus
# 1) eemalda https
# 2) eemalda confutil
#  awk '/e-gov.*e-gov/ { print $0; }' | \
#  vali e-gov -> e-gov (serva eemaldamiseks $1, $3)
# 4) eemalda eesliide github...
# 5) eemalda (* ja )

# Märkmed
# awk ja sed programmides kasuta ülakomasid. Ülakomadevahelist teksti tõlgendab
# Bash literaalina. Ülakoma enda kasutamiseks ülakomade vahel kasuta '\''.
#
# sed-is on vaja paostada sümbolid $.*[\^. (){}+?| ei ole vaja paostada.
# Vt: https://unix.stackexchange.com/questions/32907/what-characters-do-i-need-to-escape-when-using-sed-in-a-sh-script
#
# Jätkamismärgid \ peavad olema reas viimased.
