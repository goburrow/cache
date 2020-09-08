# Cache performance report

Run all tests
```
./report.sh
```

Run individual test
```
go test -v -run Wikipedia
./visualize-request.sh request_wikipedia-*.txt
./visualize-size.sh size_wikipedia-*.txt
open out.png
```

## Traces

Name         | Source
------------ | ------
Address      | [University of California, San Diego](http://cseweb.ucsd.edu/classes/fa07/cse240a/project1.html)
CPP          | Authors of the LIRS algorithm - retrieved from [Cache2k](https://github.com/cache2k/cache2k-benchmark)
Glimpse      | Authors of the LIRS algorithm - retrieved from [Cache2k](https://github.com/cache2k/cache2k-benchmark)
Multi2       | Authors of the LIRS algorithm - retrieved from [Cache2k](https://github.com/cache2k/cache2k-benchmark)
OLTP         | Authors of the ARC algorithm - retrieved from [Cache2k](https://github.com/cache2k/cache2k-benchmark)
ORMBusy      | GmbH - retrieved from [Cache2k](https://github.com/cache2k/cache2k-benchmark)
Sprite       | Authors of the LIRS algorithm - retrieved from [Cache2k](https://github.com/cache2k/cache2k-benchmark)
Wikipedia    | [WikiBench](http://www.wikibench.eu/)
YouTube      | [University of Massachusetts](http://traces.cs.umass.edu/index.php/Network/Network)
WebSearch    | [University of Massachusetts](http://traces.cs.umass.edu/index.php/Storage/Storage)
