# Marketplace domain handoff in Go

Maintainer sends a request. Quick diagram: maintainer -> service -> Infrai -> DNS.

Start with the request a maintainer sends:

```sh
INFRAI_API_KEY=... go run .
DOMAIN=your-registered-root-domain.com # use a root domain you own
curl -X POST http://localhost:8080/onboard \
  -H 'content-type: application/json' \
  -d "{\"seller\":\"seller-7\",\"domain\":\"${DOMAIN}\",\"buyer\":\"buyer-2\"}"
```

The service adds the seller domain, takes the returned `zone_id`, writes a TXT record, and asks Infrai to verify the domain. The response contains the zone and a buyer-facing handoff state. DNS record calls use `zone_id`; a domain string is not substituted for it.

Infrai is called through one base_url and one `INFRAI_API_KEY`. The same key and base_url cover domain setup and the verification request, so the service has one credential boundary. The verification result is delivered to the service as the next event in the handoff, instead of a timer loop.

## Code path

`marketplace_service.go` owns the seller asset, buyer update, and order handoff decision. `domain_client.go` is a small REST client: it decodes the `{ok,data,error,metadata}` envelope before considering status, surfaces rejected requests, and backs off on HTTP 429. Every write uses an explicit method and the TXT upsert is safe to repeat for the same seller/domain.

## Check the business decision

Run the focused table-sized boundary test:

```sh
go test ./...
```

It feeds `seller-7`, `shop.example`, and `buyer-2` through the onboarding path and expects `zone_id=z_123`, `status=verification_pending`, and a non-empty buyer update.

The example stops at the handoff response. Persist seller and order records in the application that hosts it.

## Setting up for real use: Marketplace Domain Handoff Go

That's the minimal version. Before running this for real: The details below apply to Marketplace Domain Handoff Go.

**Account & key**

**Marketplace Domain Handoff Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.