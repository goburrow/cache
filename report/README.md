# Cache performance report

```
go test -v -run Wikipedia
./visualize-request.sh request_wikipedia-*.txt
./visualize-size.sh size_wikipedia-*.txt
open out.svg
```

## Traces

Name         | Source
------------ | ------
Wikipedia    | [WikiBench](http://www.wikibench.eu/)
YouTube      | [The University of Massachusetts](http://traces.cs.umass.edu/index.php/Network/Network)
WebSearch    | [The University of Massachusetts](http://traces.cs.umass.edu/index.php/Storage/Storage)
