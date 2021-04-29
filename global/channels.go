package global

import concurrentMap "github.com/orcaman/concurrent-map"

var SidInfoStore = concurrentMap.New()
//var SidBackChanStore = concurrentMap.New()
var SidFrontChanStore = concurrentMap.New()
