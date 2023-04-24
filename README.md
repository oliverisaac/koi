# koi

Koi is a wrapper around kubectl that provides additional features:


#### -x shorthand flag for --context

#### `export` commandlet into which you can pipe `kubectl get secrets -o yaml`

#### `-o=yq` for pretty-print yaml. Also, supports `-o=jq`. 

#### `--yq=FILTER` to do output to yq and pass that filter to yq. Same applies to `--jq=FILTER`


# Installation:

```
brew install oliverisaac/tap/koi
```
