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
Address      | [University of California, San Diego](http://cseweb.ucsd.edu/classes/fa07/cse240a/project1.html)
CPP          | [Cache2k](http://cache2k.org/benchmarks.html)
Glimpse      | [Cache2k](http://cache2k.org/benchmarks.html)
Multi2       | [Cache2k](http://cache2k.org/benchmarks.html)
OLTP         | [Cache2k](http://cache2k.org/benchmarks.html)
ORMBusy      | [Cache2k](http://cache2k.org/benchmarks.html)
Sprite       | [Cache2k](http://cache2k.org/benchmarks.html)
Wikipedia    | [WikiBench](http://www.wikibench.eu/)
YouTube      | [The University of Massachusetts](http://traces.cs.umass.edu/index.php/Network/Network)
WebSearch    | [The University of Massachusetts](http://traces.cs.umass.edu/index.php/Storage/Storage)
