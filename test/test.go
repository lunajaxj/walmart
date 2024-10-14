// 使用golang 搜索其他程序内存空间的字符串地址，中文注释

package main

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func main() {
	for {
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)
		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/873221212", nil)
		//request, _ := http.NewRequest("GET", "https://www.walmart.com", nil)
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
		//request.Header.Add("Cookie", "_pxvid=fbfbd315-cedb-11ed-ab0b-4a5079654d6f; vtc=X7jSdDtF3diOAw3Wb1nRD0; ACID=b5899005-cf67-454b-9705-3962b87b61c4; hasACID=true; AID=wmlspartner%3D0%3Areflectorid%3D0000000000000000000000%3Alastupd%3D1680168106877; abqme=true; s_pers_2=+s_v%3DY%7C1682392087123%3B+gpv_p11%3DWalmart%2520Photo%253A%2520LiveLink%253A%2520Homepage%7C1682392087136%3B+gpv_p44%3DWalmart%2520Photo%253A%2520LiveLink%7C1682392087138%3B+s_vs%3D1%7C1682392087142%3BuseVTC%3DY%7C1745505522; _pxhd=7cd4b17cb2ffaf65a67c8e36127571971a6e791eb36e48f12f6e972470ee5b5c:fbfbd315-cedb-11ed-ab0b-8ff566896a0f; TBV=7; akhintab=homevar; chsn_cnsnt=www.walmart.com%3AC0001%2CC0002%2CC0003%2CC0004%2CC0005; tglr_sess_count=1; pmpdid=fe7b49e3-12eb-484c-b5bb-c3cfffaa1a74; locGuestData=eyJpbnRlbnQiOiJTSElQUElORyIsImlzRXhwbGljaXQiOmZhbHNlLCJzdG9yZUludGVudCI6IlBJQ0tVUCIsIm1lcmdlRmxhZyI6ZmFsc2UsImlzRGVmYXVsdGVkIjp0cnVlLCJwaWNrdXAiOnsibm9kZUlkIjoiMzA4MSIsInRpbWVzdGFtcCI6MTY4NTc3OTk1NjI2N30sInNoaXBwaW5nQWRkcmVzcyI6eyJ0aW1lc3RhbXAiOjE2ODU3Nzk5NTYyNjcsInR5cGUiOiJwYXJ0aWFsLWxvY2F0aW9uIiwiZ2lmdEFkZHJlc3MiOmZhbHNlLCJwb3N0YWxDb2RlIjoiOTU4MjkiLCJjaXR5IjoiU2FjcmFtZW50byIsInN0YXRlIjoiQ0EiLCJkZWxpdmVyeVN0b3JlTGlzdCI6W3sibm9kZUlkIjoiMzA4MSIsInR5cGUiOiJERUxJVkVSWSIsInRpbWVzdGFtcCI6MTY4NzMxODI4NDYxNCwic2VsZWN0aW9uVHlwZSI6IkxTX1NFTEVDVEVEIiwic2VsZWN0aW9uU291cmNlIjpudWxsfV19LCJwb3N0YWxDb2RlIjp7InRpbWVzdGFtcCI6MTY4NTc3OTk1NjI2NywiYmFzZSI6Ijk1ODI5In0sInZhbGlkYXRlS2V5IjoicHJvZDp2MjpiNTg5OTAwNS1jZjY3LTQ1NGItOTcwNS0zOTYyYjg3YjYxYzQifQ%3D%3D; akehab=ipVar; userAppVersion=main-1.81.0-f787e7-0622T1507; pxcts=c4a0e821-1335-11ee-84aa-6671734a7043; akavpau_p1=1687684668~id=28c6d32626c17d74097c733f8fe2a742; bstc=aPkpXd6b_ezUjTUpKJYg-A; _pxff_cfp=1; auth=MTAyOTYyMDE4ATNImlXmBQpAE6VVbLKR8R1o80ymFT5EKrVKZDh3UxT0MkB24ShbeSihOcD7BcIhsRkbysnFQmPJp71M%2FIX60dhb0ep8fIPX%2B0GlZ9kI%2BtGj7uTMYkOt9KkGY5%2BzSLDQ767wuZloTfhm7Wk2KcjygqjPQjfEaB1WK%2FMFlTnguVXqwB60DvjjLCdEzDPYub4jmAKl3zgS%2BtqfQE%2FJhM6DmmDfLzBYGIG%2FjdCn3udL3bMUMk70P8glgOEpLOprhDfMDCcb9mgycy9jtT1uIyOBHSdeI0EBbSPBjN7FTd9R1dR9AQAZH9qqFt6mjNswP34d4cbEP3S%2BNijCMvdHHDxrlmHpqR60H0ig%2FZfPFxfRdeX%2FtSNiKXLdsHy1VlbFQ8fYmO5VuHLO%2Bii3A8wuMC3MspE5WBBdZBCyKnCQAR7o6eg%3D; locDataV3=eyJpc0RlZmF1bHRlZCI6dHJ1ZSwiaXNFeHBsaWNpdCI6ZmFsc2UsImludGVudCI6IlNISVBQSU5HIiwicGlja3VwIjpbeyJidUlkIjoiMCIsIm5vZGVJZCI6IjMwODEiLCJkaXNwbGF5TmFtZSI6IlNhY3JhbWVudG8gU3VwZXJjZW50ZXIiLCJub2RlVHlwZSI6IlNUT1JFIiwiYWRkcmVzcyI6eyJwb3N0YWxDb2RlIjoiOTU4MjkiLCJhZGRyZXNzTGluZTEiOiI4OTE1IEdlcmJlciBSb2FkIiwiY2l0eSI6IlNhY3JhbWVudG8iLCJzdGF0ZSI6IkNBIiwiY291bnRyeSI6IlVTIiwicG9zdGFsQ29kZTkiOiI5NTgyOS0wMDAwIn0sImdlb1BvaW50Ijp7ImxhdGl0dWRlIjozOC40ODI2NzcsImxvbmdpdHVkZSI6LTEyMS4zNjkwMjZ9LCJpc0dsYXNzRW5hYmxlZCI6dHJ1ZSwic2NoZWR1bGVkRW5hYmxlZCI6dHJ1ZSwidW5TY2hlZHVsZWRFbmFibGVkIjp0cnVlLCJodWJOb2RlSWQiOiIzMDgxIiwic3RvcmVIcnMiOiIwNjowMC0yMzowMCIsInN1cHBvcnRlZEFjY2Vzc1R5cGVzIjpbIlBJQ0tVUF9JTlNUT1JFIiwiUElDS1VQX0NVUkJTSURFIl19XSwic2hpcHBpbmdBZGRyZXNzIjp7ImxhdGl0dWRlIjozOC40NzQ2LCJsb25naXR1ZGUiOi0xMjEuMzQzOCwicG9zdGFsQ29kZSI6Ijk1ODI5IiwiY2l0eSI6IlNhY3JhbWVudG8iLCJzdGF0ZSI6IkNBIiwiY291bnRyeUNvZGUiOiJVU0EiLCJnaWZ0QWRkcmVzcyI6ZmFsc2V9LCJhc3NvcnRtZW50Ijp7Im5vZGVJZCI6IjMwODEiLCJkaXNwbGF5TmFtZSI6IlNhY3JhbWVudG8gU3VwZXJjZW50ZXIiLCJpbnRlbnQiOiJQSUNLVVAifSwiaW5zdG9yZSI6ZmFsc2UsImRlbGl2ZXJ5Ijp7ImJ1SWQiOiIwIiwibm9kZUlkIjoiMzA4MSIsImRpc3BsYXlOYW1lIjoiU2FjcmFtZW50byBTdXBlcmNlbnRlciIsIm5vZGVUeXBlIjoiU1RPUkUiLCJhZGRyZXNzIjp7InBvc3RhbENvZGUiOiI5NTgyOSIsImFkZHJlc3NMaW5lMSI6Ijg5MTUgR2VyYmVyIFJvYWQiLCJjaXR5IjoiU2FjcmFtZW50byIsInN0YXRlIjoiQ0EiLCJjb3VudHJ5IjoiVVMiLCJwb3N0YWxDb2RlOSI6Ijk1ODI5LTAwMDAifSwiZ2VvUG9pbnQiOnsibGF0aXR1ZGUiOjM4LjQ4MjY3NywibG9uZ2l0dWRlIjotMTIxLjM2OTAyNn0sImlzR2xhc3NFbmFibGVkIjp0cnVlLCJzY2hlZHVsZWRFbmFibGVkIjp0cnVlLCJ1blNjaGVkdWxlZEVuYWJsZWQiOnRydWUsImFjY2Vzc1BvaW50cyI6W3siYWNjZXNzVHlwZSI6IkRFTElWRVJZX0FERFJFU1MifV0sImh1Yk5vZGVJZCI6IjMwODEiLCJpc0V4cHJlc3NEZWxpdmVyeU9ubHkiOmZhbHNlLCJzdXBwb3J0ZWRBY2Nlc3NUeXBlcyI6WyJERUxJVkVSWV9BRERSRVNTIl0sInNlbGVjdGlvblR5cGUiOiJMU19TRUxFQ1RFRCJ9LCJyZWZyZXNoQXQiOjE2ODc4Mzk5OTQ5OTQsInZhbGlkYXRlS2V5IjoicHJvZDp2MjpiNTg5OTAwNS1jZjY3LTQ1NGItOTcwNS0zOTYyYjg3YjYxYzQifQ%3D%3D; assortmentStoreId=3081; hasLocData=1; mobileweb=0; xpth=x-o-mart%2BB2C~x-o-mverified%2Bfalse; xpa=0Iadf|0uTG6|1JRNS|4CECO|5e9Fg|8cYMq|8oGja|APtyx|BukPC|CLzBA|DcdL-|Du0vv|Edk-I|GmDfi|GycPV|H-URg|IedF3|IhmrE|KvYZX|LbWEb|MT_LO|Sd-TJ|TeIx2|Uwlt9|VyZuz|WvDEF|YnYws|ZllwQ|Zwl71|_pFGX|a_rrh|b420n|c-Etr|cksJm|dfrMM|fFWC6|gjQMr|ikhNy|jyp9o|kpr0y|nOUgH|o7U1C|ox1K8|pPaFS|pcXyb|qi56T|rldce|v4Ppy|vTjpJ|wXvq0; exp-ck=0Iadf10uTG618cYMq1APtyx1BukPC1Du0vv1GmDfi2GycPV1H-URg2KvYZX1LbWEb1Sd-TJ2TeIx22Uwlt91WvDEF1YnYws4_pFGX1a_rrh2b420n1cksJm1dfrMM1ikhNy1kpr0y1o7U1C1pcXyb1qi56T1rldce1v4Ppy2vTjpJ1; ak_bmsc=D57F9AA71285EC670CF5AF654268629B~000000000000000000000000000000~YAAQMPAgF5g4lfKIAQAA2Yfj+hSJZZVo63y/SEQeRxmB8HDvBbQDtST9BXrOE5ioyuJwr67guoV3RPqP6PWurTgJXOo0Kjo2wLrGxRf+yNd8Qzcpbohv6DKXv025SylA5XWD2q6FIVKHU7Nn3b4mhcGN+7smkjmxdQQ8f0qkk16ZjL/h5u3GIdzqtftY5abPE+Ko1xeqPDqKnHj4PldsLLUMmdqcGw5DoMqGOcD5SvISPJNZA+uD7CQyopfiGi877dfwOnnyT2SX10TPqoElXNg+e8Pu+ueoWfMmK4LPwpD1/M1RAKL/NtFLx8mMt2BaplmOID48X3X7/Wv9bYECfioA8ZPTsDpdzYMzSSOwgk0XdGItCPASA8xgZkpK4ORqRjzdK27ekzCqlJJFeEBcR65/l66I+Fi+rCHKLdksnc21WXoNwvxgGgsKuFoNwDAhI3V+t57rfkwabCwZ4cYOrrsnjh3FQi+9SZCVLQdYH+t1o2g+8mlCJWy5pK5HyzKrVlXgN82yaQ7wklpLA+y81A==; adblocked=true; xptc=assortmentStoreId%2B3081; xpm=1%2B1687836396%2BX7jSdDtF3diOAw3Wb1nRD0~%2B0; _astc=812691fcd72913be2463f072a251d6da; com.wm.reflector=\"reflectorid:0000000000000000000000@lastupd:1687836405000@firstcreate:1687836402688\"; _px3=bde233a82db833afe0fc14667695d263c0565a2607d6e9cace99e18894747e43:gBPJrEtzjyuYIh/KeN1v0xniCmvCMYp7yW+ivOrmISmzc1QRc110W94ty9DyLQg3d6vo+Gncxo6QPTaulQdPQw==:1000:YAtbSGowcaQyL7pGgn+v8ncaGLp/V8G8/KTOxhtWHu9AGoVjNhU9ZV2OCnsCxZb+oI4FaSWJmEhXkQydaYMv+jPkwcYZzmx+Eu8N6vApcAoT2X0AyquyL9AU9TqzCsyGlYS6fhtuuaaa7FeEBdIdXCteiEpE4sI46SVEZj6SAmTWtBuJ04oK2Buc6Q3YvPoFRgJRhy35nH+N03GyrUfBeg==; _pxde=6c7e01a2feb53e12715ed6fa3aba6a6165d7887674defe2f5592da3fbb6b36f6:eyJ0aW1lc3RhbXAiOjE2ODc4MzY0MDgxNjB9; xptwg=1680494209:2547B3F7A7A1C00:5EDD771:28FF7A62:9FABB4ED:C3155032:; bm_mi=07E96A884E72431C4384C37465D5C35A~YAAQMPAgF8I+lfKIAQAAXMnj+hSFsqpIrzlH0x2D8RprrsCnqK6yFgFIl2ATMcd01X0vpi5LcuKH4mwYGFqtObGQfIpYOYez2bZ0sUt6ZXhqbRSt2/We/t0jeKB300RfePZjWkiCaBc7GkwfpthfXHtMnJuCWLLKZ5/nTgFBQGDqjhwyf+Z3sGfUDa7/t5D0L/8jFUo3K5Lc3oA/g/fC0360Sf2lKYUBFYwlWAqGNvlypiKszQSapAONc1N/DSrre8V6/OGAsudIzP8Xm2D2gbfJ+pwbxnD17nbMUN621YQBjV6ZGNqwYlohvGoFA6FqYH34gODnN5UCTB44tIBihJO3Nh0ZIpuP~1; xptwj=qq:af201efe83b12b590c38:uvO1GngAgPMCRFq5ejt9sB9+elIfhO3l65vNpX60SbHuaOznpBPESIfZfYohRGl/kom0UgUw5zfNtzdauLCcjz4Rte98lYCGGz3NmlWGgudtwWJN86lOS6xebSCFKO4n5lEuemtna33eEZFGX8RVweP6jWyZ; TS012768cf=0178545c90d10c5cfbca6d790f33197f95dbb17886986d6cfdb3354abf9f8e0a6d1872cc13e5b9580660ce0eca58468d049f513cd7; TS01a90220=0178545c90d10c5cfbca6d790f33197f95dbb17886986d6cfdb3354abf9f8e0a6d1872cc13e5b9580660ce0eca58468d049f513cd7; TS2a5e0c5c027=0881c5dd0aab20001cb6d20652707ee95ba3663f3a2de1de0182d3c34eae5874c8fef74377d5179108fe62a6ad11300010eacd8001862143f995e2b0f33a1a4aaf946d7d7d295fd460f119ecdb7a7bc5767ab740efc2f3dd5b7d9884949ba3c6; akavpau_p2=1687837018~id=cd281b9a512b3c3c373495597e2b5a13; bm_sv=6E50540E95F8F655D5B9A156D6EB8AB3~YAAQK/AgFwoEV8eIAQAAN+Lj+hQAVarca1JuHornes4PDGjNkUCGrbAocZ+ABeaBNcpemabZloQ3NT9iKObzQB1s4kVJDJZGfUAyQrmG/04qA/lPeW7tC1KPCZT07CE+4Rcp7WXPmIW8cX9rhHIym8RmU8j8xGp0bSRu4I21Cm23z96segelDCZgCusRbaHk9HRMOl3o7bK38rjXWOdRj/JTz89FQj2O63m+xmRbCYXAV2h9qzTXVGu3bB2oXrYTO9o=~1") //使用gzip压缩传输数据让访问更快
		//request.Header.Add("Cookie", "TBV=7; adblocked=false; auth=MTAyOTYyMDE4q1%2FD4sdt5RyDbHAT5QlmR%2Bm4ntu2Ku9vhbwYb%2FWBBYIE%2BcM9jHdlU%2B8W%2BdIgg3shGk2GRRMwDTOGD2B1kGp16DpV%2FGmhPBpRGsRJiMxoymIPYNKD3r%2Bl9JxCjKuUZQAL767wuZloTfhm7Wk2KcjygobRHThsmZk%2BGcqTfIab85RO7STmNzaNF6B56NEObaVpT9T6V2hcsKRc9yGId5lDvc025WCkV3eM5drwawMM4WAUMk70P8glgOEpLOprhDfMDCcb9mgycy9jtT1uIyOBHTnTWB%2Fj%2BucpSCIMgvbPEX8u%2Flf7ffAcNCC1ljCPfqmuDnnaXsxaV3%2BCNsBfRPeOoaMnLb7NiP8pX5VQQshwZk3%2FtSNiKXLdsHy1VlbFQ8fYg%2Fyq1H0EkPNmbbp5Ssx9F0jyrOXbKKhH072NS%2FW0j%2FU%3D; ACID=3037b693-8a7b-42d9-bd14-3145254cbff5; hasACID=true; locDataV3=eyJpc0RlZmF1bHRlZCI6dHJ1ZSwiaXNFeHBsaWNpdCI6ZmFsc2UsImludGVudCI6IlNISVBQSU5HIiwicGlja3VwIjpbeyJidUlkIjoiMCIsIm5vZGVJZCI6IjMwODEiLCJkaXNwbGF5TmFtZSI6IlNhY3JhbWVudG8gU3VwZXJjZW50ZXIiLCJub2RlVHlwZSI6IlNUT1JFIiwiYWRkcmVzcyI6eyJwb3N0YWxDb2RlIjoiOTU4MjkiLCJhZGRyZXNzTGluZTEiOiI4OTE1IEdlcmJlciBSb2FkIiwiY2l0eSI6IlNhY3JhbWVudG8iLCJzdGF0ZSI6IkNBIiwiY291bnRyeSI6IlVTIiwicG9zdGFsQ29kZTkiOiI5NTgyOS0wMDAwIn0sImdlb1BvaW50Ijp7ImxhdGl0dWRlIjozOC40ODI2NzcsImxvbmdpdHVkZSI6LTEyMS4zNjkwMjZ9LCJpc0dsYXNzRW5hYmxlZCI6dHJ1ZSwic2NoZWR1bGVkRW5hYmxlZCI6dHJ1ZSwidW5TY2hlZHVsZWRFbmFibGVkIjp0cnVlLCJodWJOb2RlSWQiOiIzMDgxIiwic3RvcmVIcnMiOiIwNjowMC0yMzowMCIsInN1cHBvcnRlZEFjY2Vzc1R5cGVzIjpbIlBJQ0tVUF9JTlNUT1JFIiwiUElDS1VQX0NVUkJTSURFIl0sInNlbGVjdGlvblR5cGUiOiJERUZBVUxURUQifV0sInNoaXBwaW5nQWRkcmVzcyI6eyJsYXRpdHVkZSI6MzguNDgyNjc3LCJsb25naXR1ZGUiOi0xMjEuMzY5MDI2LCJwb3N0YWxDb2RlIjoiOTU4MjkiLCJjaXR5IjoiU2FjcmFtZW50byIsInN0YXRlIjoiQ0EiLCJjb3VudHJ5Q29kZSI6IlVTIiwibG9jYXRpb25BY2N1cmFjeSI6ImxvdyIsImdpZnRBZGRyZXNzIjpmYWxzZX0sImFzc29ydG1lbnQiOnsibm9kZUlkIjoiMzA4MSIsImRpc3BsYXlOYW1lIjoiU2FjcmFtZW50byBTdXBlcmNlbnRlciIsImludGVudCI6IlBJQ0tVUCJ9LCJpbnN0b3JlIjpmYWxzZSwiZGVsaXZlcnkiOnsiYnVJZCI6IjAiLCJub2RlSWQiOiIzMDgxIiwiZGlzcGxheU5hbWUiOiJTYWNyYW1lbnRvIFN1cGVyY2VudGVyIiwibm9kZVR5cGUiOiJTVE9SRSIsImFkZHJlc3MiOnsicG9zdGFsQ29kZSI6Ijk1ODI5IiwiYWRkcmVzc0xpbmUxIjoiODkxNSBHZXJiZXIgUm9hZCIsImNpdHkiOiJTYWNyYW1lbnRvIiwic3RhdGUiOiJDQSIsImNvdW50cnkiOiJVUyIsInBvc3RhbENvZGU5IjoiOTU4MjktMDAwMCJ9LCJnZW9Qb2ludCI6eyJsYXRpdHVkZSI6MzguNDgyNjc3LCJsb25naXR1ZGUiOi0xMjEuMzY5MDI2fSwiaXNHbGFzc0VuYWJsZWQiOnRydWUsInNjaGVkdWxlZEVuYWJsZWQiOnRydWUsInVuU2NoZWR1bGVkRW5hYmxlZCI6dHJ1ZSwiYWNjZXNzUG9pbnRzIjpbeyJhY2Nlc3NUeXBlIjoiREVMSVZFUllfQUREUkVTUyJ9XSwiaHViTm9kZUlkIjoiMzA4MSIsImlzRXhwcmVzc0RlbGl2ZXJ5T25seSI6ZmFsc2UsInN1cHBvcnRlZEFjY2Vzc1R5cGVzIjpbIkRFTElWRVJZX0FERFJFU1MiXSwic2VsZWN0aW9uVHlwZSI6IkRFRkFVTFRFRCJ9LCJyZWZyZXNoQXQiOjE2ODc4NDA0MTk4MjUsInZhbGlkYXRlS2V5IjoicHJvZDp2MjozMDM3YjY5My04YTdiLTQyZDktYmQxNC0zMTQ1MjU0Y2JmZjUifQ%3D%3D; assortmentStoreId=3081; hasLocData=1; locGuestData=eyJpbnRlbnQiOiJTSElQUElORyIsImlzRXhwbGljaXQiOmZhbHNlLCJzdG9yZUludGVudCI6IlBJQ0tVUCIsIm1lcmdlRmxhZyI6ZmFsc2UsImlzRGVmYXVsdGVkIjp0cnVlLCJwaWNrdXAiOnsibm9kZUlkIjoiMzA4MSIsInRpbWVzdGFtcCI6MTY4NzgzNjgxOTgyMiwic2VsZWN0aW9uVHlwZSI6IkRFRkFVTFRFRCJ9LCJzaGlwcGluZ0FkZHJlc3MiOnsidGltZXN0YW1wIjoxNjg3ODM2ODE5ODIyLCJ0eXBlIjoicGFydGlhbC1sb2NhdGlvbiIsImdpZnRBZGRyZXNzIjpmYWxzZSwicG9zdGFsQ29kZSI6Ijk1ODI5IiwiY2l0eSI6IlNhY3JhbWVudG8iLCJzdGF0ZSI6IkNBIiwiZGVsaXZlcnlTdG9yZUxpc3QiOlt7Im5vZGVJZCI6IjMwODEiLCJ0eXBlIjoiREVMSVZFUlkiLCJ0aW1lc3RhbXAiOjE2ODc4MzY4MTk4MjEsInNlbGVjdGlvblR5cGUiOiJERUZBVUxURUQiLCJzZWxlY3Rpb25Tb3VyY2UiOm51bGx9XX0sInBvc3RhbENvZGUiOnsidGltZXN0YW1wIjoxNjg3ODM2ODE5ODIyLCJiYXNlIjoiOTU4MjkifSwidmFsaWRhdGVLZXkiOiJwcm9kOnYyOjMwMzdiNjkzLThhN2ItNDJkOS1iZDE0LTMxNDUyNTRjYmZmNSJ9; abqme=true; mobileweb=0; xpth=x-o-mart%2BB2C~x-o-mverified%2Bfalse; xpa=0KA3-|0yLUb|3K-KJ|3caIW|4CECO|5M-UN|6aedV|8ibpT|9-5I8|AkcZg|BUFSx|BukPC|DcdL-|GeS9c|GmDfi|GycPV|IedF3|J59K1|KfYac|KvYZX|Md9jg|MhDyj|OFXXb|SoVwe|U3NAT|UO2cT|Uqh-a|VDAFc|X78hm|XVDYZ|Yjred|_WAjU|_nQap|_uNDy|cksJm|dayNl|fFWC6|j1UKn|jUBiS|lFfuH|o7U1C|pyVOq|q7xXt|qi56T|u2iCd|urRIv|v4Ppy|wMBTP|wXvq0|xTsTj; exp-ck=0yLUb23caIW25M-UN16aedV18ibpT19-5I81AkcZg1BukPC1GeS9c2GmDfi2GycPV1KvYZX1UO2cT1Uqh-a1VDAFc1XVDYZ1_nQap2_uNDy1cksJm1j1UKn1lFfuH1o7U1C1qi56T1u2iCd1v4Ppy2; _pxhd=86804268494083578851e9b3e53681cc8ee4b68df7b177ac78c205f3ba63f98b:684e9f16-149b-11ee-9d4a-537c23903f29; vtc=SHnA5uOkgiPAxBqs_Q6-fs; bstc=SHnA5uOkgiPAxBqs_Q6-fs; pxcts=68d17858-149b-11ee-86f8-545641434767; _pxvid=684e9f16-149b-11ee-9d4a-537c23903f29; xptc=assortmentStoreId%2B3081; xpm=7%2B1687836819%2BSuTuBcKxHvynfXravWQuRc~%2B0; AID=wmlspartner%3D0%3Areflectorid%3D0000000000000000000000%3Alastupd%3D1687836995534; userAppVersion=main-1.81.0-f787e7-0622T1507; _astc=812691fcd72913be2463f072a251d6da; QuantumMetricSessionID=f4503cf56d30d2268f0cf767872b2ed3; QuantumMetricUserID=ecc49a6934238f58d2ac3d361d0d3c5e; xptwj=qq:e9a993a658e02f0dd6fd:mqazeyJ3SQK77p19Ges56j083/V0hcy9zlfNalDriHXRLi97KW40lf0Qq2xfy3JEQCM8NryVb4cu2IXDSUBt48jn0FU6c0oi7IOghEPhzR/l/vJu9xjRt74BK1OpR4JaF0CuVx6FANejU07WbqDolTYXcWEv5QU=; bm_mi=AAB968FE5B5827CB2F1E76C2EA8D0351~YAAQFJHIF5kky92IAQAA9wfy+hQYTV4bhQtTnRvpOyZ1bPFhpsmYYKwifRDx3OGPtSGpJPywWCC8t2wLTrm1Oew6Ax4ju6FyslMBOl01rr0fOPGOhiL5fTWcnnwR0JCy0ULKd+KGnn8AQPHNtrUM1/Geyg6Qgv8yv2XHQJmQM8Ziax429ZSNLr09cQxh1tRmd5wa8IAP92znIOm2pud7VUje672Y3ONPQ+m/71y9OaQPEp5OGyBi8NlVC9qJOfRDaGcSss2dUDijTT/xNjjd59KQlpwuHxZ3jwOvQRbf1ZSkUevoQ+4NKmrDKoLdXk+ZunsNxKuE9Lv9PQr2XKPCYeSPzlN1wPMugmUsf6qLc0Rlv9mfZUgPnXqAzDyWnjwX8HnGTOU7ZqVgQ4g/6RYtZbfZbGrjo2R+Eg==~1; ak_bmsc=0CD734FBE08623E61667D3723D0D57DC~000000000000000000000000000000~YAAQFJHIF5wky92IAQAA1wjy+hQUW5pged2cK4dI/plu8KSOqXY75VnCUF1r6rnn103sXO6Lx+2db76px5C01WOuyPN2yuI5uF/SX5DyzFL+mH/Onr4JNqLV7G0BctGB/nGaIil/3X5RWCnFzuyhUro0Y7PEeqpw3EY4Gsb+62RdxLliKLv4Qf83SuOf698K+sgYjx1yD6grJ7zNg4iPdXcUSn3VvG6dvNt5xeP2RNacdgsVpMS5hicI/3z3skA5GcFIsZ6pG7k/AVbHjyBJ68wk8NINrwIJwS0wITABlvwBsLldBjFXSI22G60XGYgOsBXN9yXKp195G4pUqy7BiNFczR7Ist0ZJ6XylIMPzw+OkjPsI0dvjkK1S732c2H2TihLP+SYBJXM6MI9+1tmJaMtdZrjz4nhuhdOqaN1YS4IQ+UzEYC4Uf5wA8gPW3cgWrxZSz7/B6Av4s63DE+hby6hj1ONvTtIWxXN/dIhmum56GyPYkklGPE8Q1vXMYxnVcGWDcT5NW9wJ3duMIoeNyvqM6rjujnW6BHWqFOnwvtVFcqTKQkYaN8GHow0eupHmPM3zXOVCY77b9qj3HWWMkIUBmkvSpxV3e26C12uVy/VAhALwiC7Oadr; com.wm.reflector=\"reflectorid:0000000000000000000000@lastupd:1687837347000@firstcreate:1687836995534\"; _pxff_cfp=1; akavpau_p2=1687837947~id=dd69ff34901788fc4c3e7a43139065d6; xptwg=2951964868:95579F7D5B74D0:17C065F:7605F696:9429418A:1644603D:; TS012768cf=016ea84bd27aa96614eb32b6028ea52dc20ed16f5e23de5e82d3390ac6f8bade4c7fd6f98fe3a343fac0277f4cc3dac3ff27ea1c61; TS01a90220=016ea84bd27aa96614eb32b6028ea52dc20ed16f5e23de5e82d3390ac6f8bade4c7fd6f98fe3a343fac0277f4cc3dac3ff27ea1c61; TS2a5e0c5c027=08039d86dcab20002d4ee1103185ffb899e10cdbcf68c67ac9a77ae2e186e70f875014edb5886e7b086cf6ab711130001616c6014d742c1c61f6f5935275862654879783630e377f83db3acac48d7ee54c91c99a9d426dc318baede10500b791; bm_sv=945BF89552022C40BAF49EA190373A62~YAAQFJHIF64ky92IAQAAJQ/y+hQvpgjmcGF+HHQlYdenVpLUNMcFxAx19Pp7xXBuOfWB7RAhK7VXPiYmA3iZ2zAEMF5CyKuFmbfuzlUzfLZOLgJyhGon3IzUXU5y67j5ksAXvd4NAULuuYuFxLYJc3LOlmPVz5EfOIA9B7ZlBNhB33SDZRwR+CR3CAoyKxIrt9QLn4RbK0oQi4bObx0jOn8CnpJTWjkl58U50WbDTEEuVmpsNXDYtccEFOqx1lq4Cpk=~1; _px3=8fb322c52f6c32cb619a01c8c3489e34a6e03f3905990e27e94664c39150395b:vxldb4r7zFMicdTQn3uyKn1m6WDxlQ5PvbYyUmT8WNKlIJO0/U45T7SD29XHje+vLtXpdSYh/8bylrAdo9D9Fw==:1000:ehgleMfYSprtTW7/WjIi5tI4ErLu7s7JB9GBgAAFP0V7gdxkHerpMIzoGFmhCYP7qP2O2IlPEKHBcvB/q9QAxoJUcH0Ej9Zm/2oDYFu9pVfgOgzDt2mW4qveGbdwrV+sOgJJCI6GNjmqLBI5u0AkeU+tfCgRahYAWzszJv0Toiiz9DcgIew3PW7rTodbkiM2q0o3ezCrcSBXULAX0Qgq9g==; _pxde=ec04ce4148043aa46bef525aece9c9aa867cb6531710820a0a6b8ab80f089336:eyJ0aW1lc3RhbXAiOjE2ODc4MzczNDk3ODN9")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("Accept-Language", "zh")
		request.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Google Chrome";v="114"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `none`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")

		response, err := client.Do(request)
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：")
				continue
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)
				continue
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)
				continue
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：")
				continue
			}
		}
		result := ""
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body) // gzip解压缩
			if err != nil {
				log.Println("解析body错误，重新开始")
				continue
			}
			defer reader.Close()
			con, err := io.ReadAll(reader)
			if err != nil {
				log.Println("gzip解压错误，重新开始")
				continue
			}
			result = string(con)
		} else {
			dataBytes, err := io.ReadAll(response.Body)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
					log.Println("代理IP无效，自动切换中")
					log.Println("连续出现代理IP无效请联系我，重新开始")
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始")
				}
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：")
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：")
			}
			continue

		}
		defer response.Body.Close()
		//result := string(dataBytes)
		//fmt.Println(result)
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			fmt.Println("被风控")
			continue
		}
		title := regexp.MustCompile("\"productName\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
		if len(title) == 0 {
			//log.Println("品牌获取失败id："+id)
		} else {
			log.Println(title[0][1])
		}
		fmt.Println("无风控")
	}

}
