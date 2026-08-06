# Google Cloud AAPL × DeltaSignal MCP client demo

**[Open the professional public demo](https://aitrailblazer.github.io/deltasignal-ai-agent/client-demo/google-cloud-aapl/)**

This reference implementation shows how a client can run bounded Google Cloud
agents over DeltaSignal MCP while preserving filing provenance, point-in-time
boundaries, ATLAS-7 applicability, missing evidence, and access status.

Start with the visual guide:

- [Client process and runnable demo](Google_Cloud_AAPL_DeltaSignal_MCP_Client_Demo_2026_08_06.html)
- [Daily backend + Google agents architecture](../../docs/DeltaSignal_Daily_Backend_to_Google_Agents_Architecture_2026_08_06.html)
- [Architecture PNG](DeltaSignal_Daily_Backend_to_Google_Agents_Architecture_2026_08_06.png)
- [Three-minute video scenario](../../docs/Google_Cloud_AAPL_MCP_Three_Minute_Video_Scenario_2026_08_06.html)

Run the no-network fixture from the repository root:

```bash
GOWORK=off go run ./cmd/aapl-client-demo -mode fixture
```

Live mode is opt-in, never settles a payment, and returns `access_required` on an MCP `402`.
