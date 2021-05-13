package global

import concurrentMap "github.com/orcaman/concurrent-map"

//var SidInfoStore = concurrentMap.New() // session 代替了
//var SidBackChanStore = concurrentMap.New()
var SidFrontChanStore = concurrentMap.New()
