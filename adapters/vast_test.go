package adapters_test

import (
	"testing"

	"github.com/prebid/prebid-server/adapters"
	"github.com/stretchr/testify/assert"
)

func TestHidePricing(t *testing.T) {
	adm := adapters.HidePricing(`<?xml version='1.0' encoding='UTF-8'?>
	<VAST version='2.0.1'>
	   <Ad id='Y206YWQ9MjEzNzQ2MSZjcz0yMzA3MDgxJnR1PTA'>
		  <InLine>
			 <AdSystem><![CDATA[ARSPY]]></AdSystem>
			 <AdTitle><![CDATA[Triton Digital a2x]]></AdTitle>
			 <Impression id='ImpressionConfirmationURL'><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/1?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/2?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/3?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/4?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression><Pricing currency='USD' model='CPM'>3.20</Pricing>
			 <Creatives>
				<Creative id='72745' sequence='1'>
				   <Linear>
					   <Duration>00:01:10</Duration>
				   </Linear>
				</Creative>
			 </Creatives>
		  </InLine>
	   </Ad>
	</VAST>`)

	assert.Equal(t, adm, `<?xml version='1.0' encoding='UTF-8'?>
	<VAST version='2.0.1'>
	   <Ad id='Y206YWQ9MjEzNzQ2MSZjcz0yMzA3MDgxJnR1PTA'>
		  <InLine>
			 <AdSystem><![CDATA[ARSPY]]></AdSystem>
			 <AdTitle><![CDATA[Triton Digital a2x]]></AdTitle>
			 <Impression id='ImpressionConfirmationURL'><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/1?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/2?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/3?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/4?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Creatives>
				<Creative id='72745' sequence='1'>
				   <Linear>
					   <Duration>00:01:10</Duration>
				   </Linear>
				</Creative>
			 </Creatives>
		  </InLine>
	   </Ad>
	</VAST>`)

	admNoChange := adapters.HidePricing(`<?xml version='1.0' encoding='UTF-8'?>
	<VAST version='2.0.1'>
	   <Ad id='Y206YWQ9MjEzNzQ2MSZjcz0yMzA3MDgxJnR1PTA'>
		  <InLine>
			 <AdSystem><![CDATA[ARSPY]]></AdSystem>
			 <AdTitle><![CDATA[Triton Digital a2x]]></AdTitle>
			 <Impression id='ImpressionConfirmationURL'><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/1?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/2?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/3?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/4?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Creatives>
				<Creative id='72745' sequence='1'>
				   <Linear>
					   <Duration>00:01:10</Duration>
				   </Linear>
				</Creative>
			 </Creatives>
		  </InLine>
	   </Ad>
	</VAST>`)

	assert.Equal(t, admNoChange, `<?xml version='1.0' encoding='UTF-8'?>
	<VAST version='2.0.1'>
	   <Ad id='Y206YWQ9MjEzNzQ2MSZjcz0yMzA3MDgxJnR1PTA'>
		  <InLine>
			 <AdSystem><![CDATA[ARSPY]]></AdSystem>
			 <AdTitle><![CDATA[Triton Digital a2x]]></AdTitle>
			 <Impression id='ImpressionConfirmationURL'><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/1?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/2?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/3?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Impression><![CDATA[http://mock.live.streamtheworld.com/track/audio/impression/4?lsid=cookie:generated-by-arspy&cb=1565288161382]]></Impression>
			 <Creatives>
				<Creative id='72745' sequence='1'>
				   <Linear>
					   <Duration>00:01:10</Duration>
				   </Linear>
				</Creative>
			 </Creatives>
		  </InLine>
	   </Ad>
	</VAST>`)
}
