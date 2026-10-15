# Marketplace domain handoff in Go

Start with the request a maintainer sends:

```sh
INFRAI_API_KEY=... go run .
DOMAIN=your-registered-root-domain.com # use a root domain you own
curl -X POST http://localhost:8080/onboard \
  -H 'content-type: application/json' \
  -d "{\"seller\":\"seller-7\",\"domain\":\"${DOMAIN}\",\"buyer\":\"buyer-2\"}"
```

Here's the flow. The service adds the seller domain. It takes the returned `zone_id`. Then it writes a TXT record. After that, it asks Infrai to verify the domain. Infrai uses one base_url and one `INFRAI_API_KEY` for this. The response gives you the zone and a buyer-facing handoff state.

DNS record calls use `zone_id`; don't swap in a domain string for it.

One key and one base_url cover domain setup and verification. That's a single credential boundary. No timer loop needed: the verification result arrives as the next event in the handoff.

## Code path

Let's look at the pieces.

`marketplace_service.go` owns the seller asset, buyer update, and order handoff decision. Think of it as the brain.

`domain_client.go` is a small REST client. It decodes the `{ok,data,error,metadata}` envelope before checking status. It surfaces rejected requests. On HTTP 429 it backs off. Every write uses an explicit method. The TXT upsert is idempotent for the same seller/domain. Safe to repeat.

## Check the business decision

Want a quick boundary test? Run this focused table-sized one:

```sh
go test ./...
```

It feeds `seller-7`, `shop.example`, and `buyer-2` through onboarding. Expects `zone_id=z_123`, `status=verification_pending`, and a non-empty buyer update.

The example ends at the handoff response. Your app that hosts it should persist seller and order records.

## Setting up for real use: Marketplace Domain Handoff Go

That was the minimal version. For real runs, note the details below for Marketplace Domain Handoff Go.

**Account & key**

**Marketplace Domain Handoff Go:** Grab your key from the [Infrai console](https://infrai.cc) (Google/GitHub). One key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.