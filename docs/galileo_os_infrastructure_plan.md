**INFRASTRUCTURE BLUEPRINT**

**Galileo OS**

*The Brain of Businesses*

A complete, polyglot, open-source-first technical infrastructure plan

Designed for staged delivery from MVP to enterprise GA

| **PREPARED FOR** | Emmanuel — Founder, Galileo OS (galileoos.com) |
| --- | --- |
| **DOCUMENT TYPE** | Technical infrastructure plan + phased launch roadmap |
| **PREFERRED LANGUAGES** | Go (orchestration core), Rust (edge + perf), Python (agents/ML), C (embedded), React Native (mobile) |
| **PHILOSOPHY** | Open-source first. Self-hostable. Multi-tenant from day one. Skills-driven intelligence. |
| **DELIVERY MODEL** | 5 stages over ~18 months. Each stage ships paying value, not internal milestones. |
| **DATE** | May 2026 — supersedes all prior planning notes |

# Table of Contents

- [1. Executive Summary](#1-executive-summary)
  - [The five things this plan commits to](#the-five-things-this-plan-commits-to)
  - [What success looks like at each stage](#what-success-looks-like-at-each-stage)
  - [Cost target and unit economics, summarized](#cost-target-and-unit-economics-summarized)
- [2. Architecture Overview](#2-architecture-overview)
  - [2.1 The eight layers, top to bottom](#21-the-eight-layers-top-to-bottom)
  - [2.2 Architecture diagram (text rendering)](#22-architecture-diagram-text-rendering)
  - [2.3 The four cross-cutting concerns](#23-the-four-cross-cutting-concerns)
  - [2.4 The v7 build discipline](#24-the-v7-build-discipline)
- [3. Day-Zero Tenant Onboarding: How a Company Becomes a Galileo Tenant](#3-day-zero-tenant-onboarding-how-a-company-becomes-a-galileo-tenant)
  - [3.1 The philosophy: onboarding is when the Brain is built](#31-the-philosophy-onboarding-is-when-the-brain-is-built)
  - [3.2 The Onboarding Crew (six agents, one workflow)](#32-the-onboarding-crew-six-agents-one-workflow)
  - [3.3 The eight-step onboarding flow](#33-the-eight-step-onboarding-flow)
  - [3.4 The operator review (the one human gate)](#34-the-operator-review-the-one-human-gate)
  - [3.5 Performance gate (pre-registered)](#35-performance-gate-pre-registered)
  - [3.6 Codebase onboarding (Stage 2+, opt-in)](#36-codebase-onboarding-stage-2-opt-in)
  - [3.7 Re-onboarding (continuous, not one-shot)](#37-re-onboarding-continuous-not-one-shot)
- [4. The Stack, Layer by Layer](#4-the-stack-layer-by-layer)
  - [4.1 Layer 0 — Identity, Tenancy, Foundation](#41-layer-0-identity-tenancy-foundation)
  - [4.2 Layer 1 — LLM Gateway and Inference](#42-layer-1-llm-gateway-and-inference)
  - [4.3 Layer 2 — Workflow and Agent Orchestration (the kernel)](#43-layer-2-workflow-and-agent-orchestration-the-kernel)
    - [Why durable execution matters for an agent OS](#why-durable-execution-matters-for-an-agent-os)
    - [Concrete pattern: how a single agent action is structured](#concrete-pattern-how-a-single-agent-action-is-structured)
  - [4.4 Layer 3 — The Company Brain (Memory + Knowledge)](#44-layer-3-the-company-brain-memory-knowledge)
    - [Three kinds of memory, one storage substrate](#three-kinds-of-memory-one-storage-substrate)
    - [Ingestion pipeline (the moment a doc enters the Brain)](#ingestion-pipeline-the-moment-a-doc-enters-the-brain)
  - [4.5 Layer 4 — Skills Library (the intelligence layer)](#45-layer-4-skills-library-the-intelligence-layer)
    - [The SKILL.md format as the standard](#the-skillmd-format-as-the-standard)
    - [Pre-built Skill packs to mirror at launch](#pre-built-skill-packs-to-mirror-at-launch)
  - [4.6 Layer 5 — Integrations and Tools](#46-layer-5-integrations-and-tools)
    - [Curated MCP server set for launch](#curated-mcp-server-set-for-launch)
    - [Custom Galileo MCP servers (we write these)](#custom-galileo-mcp-servers-we-write-these)
  - [4.7 Layer 6 — Department Modules](#47-layer-6-department-modules)
    - [Org chart model](#org-chart-model)
  - [4.8 Layer 7 — Operator Interface](#48-layer-7-operator-interface)
- [5. Language Strategy and Service Boundaries](#5-language-strategy-and-service-boundaries)
  - [5.1 Service boundary rules](#51-service-boundary-rules)
  - [5.2 The repo layout](#52-the-repo-layout)
- [6. Phased Launch Roadmap](#6-phased-launch-roadmap)
  - [6.0 How Galileo gets installed (B2B distribution surface)](#60-how-galileo-gets-installed-b2b-distribution-surface)
    - [What ships at the end of Stage 0](#what-ships-at-the-end-of-stage-0)
    - [Why Marketing first](#why-marketing-first)
    - [What ships at the end of Stage 1](#what-ships-at-the-end-of-stage-1)
    - [What ships at the end of Stage 2](#what-ships-at-the-end-of-stage-2)
    - [What ships at the end of Stage 3](#what-ships-at-the-end-of-stage-3)
    - [What ships during year 2](#what-ships-during-year-2)
  - [6.1 Rollup timeline view](#61-rollup-timeline-view)
- [7. Complete Bill of Materials](#7-complete-bill-of-materials)
- [8. Risk Register](#8-risk-register)
- [9. Build vs Buy vs Fork — Explicit Decisions](#9-build-vs-buy-vs-fork-explicit-decisions)
- [Appendix A. Curated Repo Index](#appendix-a-curated-repo-index)
  - [A.1 Layer 0 — Foundation](#a1-layer-0-foundation)
  - [A.2 Layer 1 — LLM Gateway](#a2-layer-1-llm-gateway)
  - [A.3 Layer 2 — Orchestration Kernel](#a3-layer-2-orchestration-kernel)
  - [A.4 Layer 3 — Company Brain](#a4-layer-3-company-brain)
  - [A.5 Layer 4 — Skills](#a5-layer-4-skills)
  - [A.6 Layer 5 — Integrations](#a6-layer-5-integrations)
  - [A.7 Layer 6 — Department Modules](#a7-layer-6-department-modules)
  - [A.8 Layer 7 — Operator](#a8-layer-7-operator)
  - [A.9 Cross-cutting](#a9-cross-cutting)
- [Appendix B. Stage 0 Starter Stack](#appendix-b-stage-0-starter-stack)


# 1. Executive Summary

Galileo OS is a software infrastructure layer that runs the non-engineering departments of a business — marketing, sales, customer service, support, design, accounting, HR, ops, social, content — as autonomous AI services that coordinate with each other under a single org-chart model. It is, in plain terms, the brain of a company: every business gets a Galileo, every Galileo is multi-tenant, plug-and-play, and bring-your-own-LLM.

This document is the technical infrastructure plan to build Galileo OS. It commits to specific software, specific languages, and a specific staged delivery sequence so that Galileo ships a paying product in roughly 90 days from kickoff and a credible v1.0 inside 12 months — not in two years.

## The five things this plan commits to

1. **Polyglot core, opinionated boundaries.** Go owns the orchestration kernel and all hot-path infrastructure (Temporal workers, agent runtime, MCP servers, gateway). Python owns agent business logic and ML. Rust owns the desktop shell (Tauri) and any latency-critical edge component. React Native owns the mobile companion. TypeScript owns the web admin and custom n8n nodes. C is reserved for an eventual hardware SKU and is not in scope for the first 12 months.
2. **Open-source first, self-hostable always.** Every layer of the default stack is OSS — Temporal, NATS, Supabase, LiteLLM, n8n, Chatwoot, Coolify, Opik, PostHog, Weaviate, Tauri. A Galileo customer can self-host the entire platform in their own cloud, on a single beefy VM, or on bare metal. This is a feature, not a fallback — it is how Galileo wins against ServiceNow, Salesforce, and the SaaS bundle the average startup is paying $8,000–$50,000 per year for.
3. **The Skill is the unit of intelligence, not the agent.** Following Anthropic's SKILL.md format and Tom Blomfield's 'company brain' thesis, Galileo treats the agent as a runtime and the Skill as the asset. A Skill is the encoded know-how of how a refund is processed in this company, how a sales objection is handled in this vertical, how a Q3 board update is written. Skills are versioned, evaluable, marketable, and continuously improved by an autoresearch loop in the Karpathy pattern. This is the moat.
4. **Durable execution, not fragile chains.** Every meaningful agent action runs as a Temporal workflow. That means it survives the worker dying, the LLM provider going down, the user closing the laptop, and a 14-day human approval delay. No queues, no cron, no homegrown state machines, no retries written by hand. This is how Galileo earns the right to be left alone overnight and trusted with real money.
5. **Phased launch, never a big bang.** Five stages over roughly 18 months. Each stage ships a complete, paid, useful product that one type of customer will buy on its own merits. Stage 1 ships 'Galileo for Marketing' in 12 weeks. Stage 4 ships the enterprise vertical editions in year 2. We never bet the company on a multi-year build before getting paid.
6. **Day-Zero Onboarding is the product.** A B2B customer's first hour with Galileo is not a setup wizard — it is an autonomous crawl. The Onboarding Crew (six agents, one Temporal workflow) connects to the company's existing systems read-only, walks every document, transcript, ticket, and repo, and produces the synthesized Company Brain plus a calibrated Skill manifest per department before any agent ships an output. Section 3 is the customer-facing front door of the entire product. Everything else is what runs after the front door has done its job.

## What success looks like at each stage

| **Stage** | **Window** | **Ships to** | **Headline product** |
| --- | --- | --- | --- |
| Stage 0 | Weeks 1–4 | Internal only | 'Hello Agent' — kernel boot, one demo agent runs durably with traces and budget. Mirage layer-relocation closeout committed (Reading 2 — Mirage at Layer 5, not Layer 3). |
| Stage 1 | Weeks 5–12 | Private beta (10–25 design partners) | Galileo for Marketing — Onboarding Crew GA + 5 marketing agents running as a hireable AI marketing team |
| Stage 2 | Weeks 13–24 | Public beta (open waitlist) | Galileo AI Office — adds Sales, Support, Design departments + 30 connectors + Telegram/WhatsApp gateway + Skill marketplace v1 |
| Stage 3 | Months 7–12 | GA (self-serve + sales-led) | Galileo OS 1.0 — full eight-department stack + Skill marketplace v2 with Stripe Connect + template companies (ClipMart-style) |
| Stage 4 | Year 2 | Enterprise + verticals | Galileo OS Enterprise — SOC 2 Type II, single-tenant dedicated, on-prem installer, white-label, optional vertical Skill packs (legal, real estate, healthcare, financial services) |

> **The strategic insight buried in the source material**
>
> Tom Blomfield's 'Company Brain' note in the source document is the single most important strategic input for Galileo. He writes that the bottleneck for AI automation is no longer the models — it is the lack of structured, executable domain knowledge inside companies. Skills are the answer to that bottleneck.
>
> Galileo's defensible position is therefore not 'we orchestrate agents.' Dozens of OSS projects already do that — Paperclip, gstack, Dust, AgentsMesh, mco, Mission Control, Praison, AutoGen, CrewAI. Galileo's defensible position is 'we are the place where every business's executable know-how lives, versioned and improving.' Day-Zero Onboarding is when that know-how is captured; every subsequent day is when it compounds. A Galileo tenant that has been running for six months has a Brain that no new competitor can replicate in less than six months.

## Cost target and unit economics, summarized

- Self-hosted Stage 1 deployment — single $80–$200/month VM (Hetzner CCX23 or DigitalOcean equivalent) plus pass-through LLM cost. The whole stack is designed to run on one box for the first hundred customers.
- LLM cost is metered per-tenant by LiteLLM. Each tenant has a hard monthly cap enforced at the gateway. Surprise bills are structurally impossible.
- At GA we expect a healthy SaaS gross margin (>75%) because the majority of cost variance is LLM tokens, which we charge through plus a markup on managed plans.

# 2. Architecture Overview

Galileo OS is a layered architecture. Eight horizontal layers, each with a clear responsibility, a primary technology choice, and named alternatives. The agent-as-microservice pattern runs on top of those layers. The whole stack is designed to deploy as a single Docker Compose stack at small scale and as a Kubernetes Helm chart at large scale, with no code changes between the two.

## 2.1 The eight layers, top to bottom

| **#** | **Layer** | **Owns** | **Primary tech** |
| --- | --- | --- | --- |
| 7 | Operator Interface | Web admin, mobile companion, desktop shell, Telegram/WhatsApp/Slack gateways | Next.js + Tailwind + shadcn/ui (web), React Native + Expo (mobile), Tauri + Rust (desktop) |
| 6 | Department Modules | The pre-built agent teams: Marketing, Sales, Support, Design, Finance, HR/Ops, PM, Social | Composable Skill packs over the agent runtime |
| 5 | Integrations / Tools | Outbound API surface — Gmail, Slack, Stripe, calendar, browser, social, CRM, ERP | MCP (standard) + n8n (visual) + Composio/Rube (managed catalog) |
| 4 | Skills Library | Versioned, evaluable, portable units of company know-how | Anthropic SKILL.md format + per-tenant skill registry + autoresearch loop |
| 3 | Memory / Company Brain | Vector + graph + episodic memory; ingestion of every doc, ticket, transcript | Postgres + pgvector + AGE extension; Weaviate at scale; Docling/MarkItDown/Crawl4AI/Firecrawl ingestion |
| 2 | Workflow & Agent Orchestration | Durable execution, agent graphs, role-based crews, multi-agent coordination | Temporal (Go, kernel) + LangGraph/CrewAI/Agno (agent code) + NATS JetStream (pub/sub bus) |
| 1 | LLM Gateway / Inference | Unified LLM proxy, cost metering, fallback, local inference for sensitive tenants | LiteLLM (proxy) + vLLM (production local) + Ollama (dev/edge) + Opik (tracing) |
| 0 | Identity & Tenant Foundation | Auth, multi-tenant DB, secrets, billing, audit | Supabase (Postgres+auth+realtime+storage) + Authgear (SSO) + Infisical (secrets) + Stripe Billing |

## 2.2 Architecture diagram (text rendering)

The following block diagram captures the runtime topology. Read top-down: the operator and external world interact with Galileo through layer 7; every action below it is durable, observable, and tenant-scoped.

```
┌──────────────────────────────────────────────────────────────────────────┐
│  LAYER 7  OPERATOR INTERFACE                                             │
│  Web admin (Next.js)   Mobile (React Native)   Desktop (Tauri)           │
│  Telegram bot          WhatsApp gateway        Slack app                 │
└──────────────────────────────────┬───────────────────────────────────────┘
                                   │  HTTPS / WebSocket / signed webhooks
┌──────────────────────────────────▼───────────────────────────────────────┐
│  API GATEWAY (Go)   tenant resolver, rate limit, signature verification │
└──────────┬───────────────────────────────────────────┬───────────────────┘
           │                                           │
┌──────────▼─────────────────┐               ┌─────────▼──────────────────┐
│  L6  DEPARTMENT MODULES    │               │  L5  INTEGRATIONS          │
│  Marketing  Sales  Support │  ◄───tools────┤  MCP servers  n8n  Composio│
│  Design  Finance  HR  PM   │               │  Browser-use  Playwright   │
└──────────┬─────────────────┘               └────────────────────────────┘
           │ runs as
┌──────────▼──────────────────────────────────────────────────────────────┐
│  L4  SKILLS LIBRARY    SKILL.md packs, per-tenant registry, eval suite │
└──────────┬──────────────────────────────────────────────────────────────┘
           │ executed by
┌──────────▼──────────────────────────────────────────────────────────────┐
│  L2  ORCHESTRATION KERNEL                                               │
│  Temporal Server (Go)   ◄── workflows ──   Agent Runner (Go)            │
│         │                                         │                     │
│         │          NATS JetStream  (pub/sub bus)  │                     │
│         │                                         │                     │
│         └────► LangGraph / CrewAI / Agno (Python agent code) ◄──────────┤
└──────────┬──────────────────────────────────────────┬───────────────────┘
           │                                          │
┌──────────▼─────────────┐                  ┌─────────▼──────────────────┐
│  L3  COMPANY BRAIN     │                  │  L1  LLM GATEWAY           │
│  pgvector + AGE        │                  │  LiteLLM proxy             │
│  Weaviate (at scale)   │                  │  vLLM / Ollama (local)     │
│  Docling / MarkItDown  │                  │  Opik tracing              │
│  Memary / Mem0 pattern │                  │  Per-tenant cost metering  │
└──────────┬─────────────┘                  └─────────┬──────────────────┘
           │                                          │
┌──────────▼──────────────────────────────────────────▼───────────────────┐
│  L0  TENANT FOUNDATION                                                  │
│  Supabase (Postgres + auth + RLS + realtime + storage)                  │
│  Authgear (SSO)   Infisical (secrets)   Stripe (billing)                │
│  Coolify (one-click self-host)   Helm chart (Kubernetes)                │
└─────────────────────────────────────────────────────────────────────────┘
```

## 2.3 The four cross-cutting concerns

Independent of layers, four concerns cut across everything. Each has a named owner technology so there is never ambiguity about where the logic lives.

| **Concern** | **Owner technology** | **Why this and not the alternatives** |
| --- | --- | --- |
| Observability | Opik (LLM traces) + Prometheus/Grafana (infra) + PostHog (product) | Opik is built specifically for LLM traces and works with every framework. Prometheus is the only sane choice for infra. PostHog is one tool that replaces Mixpanel + LaunchDarkly + Sentry session replay. |
| Security | Infisical (secrets) + Authgear (SSO/OIDC) + Postgres RLS + signed JWT for inter-service | Infisical is the OSS Vault that does not require a PhD to operate. Postgres Row-Level Security is the cleanest way to enforce tenant isolation in a single database — every query is automatically scoped. |
| Cost control | LiteLLM virtual keys + Temporal worker quotas + tenant-level budget caps | LiteLLM has per-key, per-tenant, per-team budget enforcement built in. We do not need to invent a billing meter — it ships with one. The Paperclip 'no surprise bills' guarantee is implemented entirely in this layer. |
| Skill quality | Promptfoo + DSPy + skill-check + nightly autoresearch loop | Promptfoo runs deterministic tests against every Skill version. DSPy auto-optimizes prompts against a metric. skill-check validates SKILL.md format. Autoresearch generates samples, judges them, patches the Skill, repeats — overnight. |

## 2.4 The v7 build discipline

Galileo OS is built the way the Alpha Sentinel v7 rebuild was run: every adoption is probed before commitment, every phase has pre-registered pass/fail gates, every closeout — including a failed phase — produces a written artifact. The goal is to keep the team from shipping things that look done but are not, and to keep ourselves honest when the gate fails. The team operates under nine rules:

1. **Probe before adopt.** Every vendor and every OSS package the kernel depends on gets an API probe or integration probe before it lands in main. The probe tests the specific behaviors the use case depends on, not the marketing copy. If the probe fails, the vendor is rejected — we do not soften the spec to accommodate a vendor we hoped would work.
2. **Pre-registered gates.** Every phase has its pass/fail thresholds locked in writing before any implementation begins. 'Stage 1 ships when X happens for at least N tenants' is decided up front, not in retrospect. This prevents the gradual gate-softening that happens when a team is tired and the deadline is real.
3. **Closeout docs for every phase, including failures.** A phase that fails its gate produces a written closeout document that names the structural finding. The Alpha Sentinel Phase 3c PEAD failure produced a document on trailing-window risk filters that informed every subsequent strategy. Galileo will run the same pattern: failed onboarding-flow experiment → CLOSEOUT_ONBOARDING_V1.md → input to v2.
4. **Calibration artifacts before implementation.** Anything that involves a threshold, a heuristic, or a tuned parameter is calibrated against real data and the calibration result is committed before the feature is built. This makes it structurally impossible to retroactively tune a threshold to make a backtest look good.
5. **Maker and checker separated.** The agent that writes code never declares it done. A second agent (or human reviewer) runs the QA pass. The CLAUDE.md and AGENTS.md files at the repo root enforce this — slash-commands like /review and /qa are non-negotiable gates before any merge.
6. **Plan before code.** Eighty percent planning and review, twenty percent execution. No keystrokes of implementation before a written plan exists. The Compound Engineering / gstack pattern adopted at the global CLAUDE.md level governs every project including Galileo itself.
7. **Compound, don't repeat.** Every bug fix, every gotcha, every 'huh that was weird' becomes a markdown file in docs/solutions/ that future agent sessions read before starting work. The cost of writing the file is ten minutes; the value compounds for years.
8. **Show me, don't tell me.** 'Tests pass' is not done. 'I clicked through the staging URL in a real browser and the new feature works end-to-end' is done. Real-browser QA via Playwright on every PR.
9. **Honest archive over softened gates.** If a Skill, a department, a connector, or a vertical edition fails its pre-registered gate, it is archived with the structural finding documented. We do not relax the gate to ship something that did not earn its way through. This is the discipline that compounds into long-term trust.

> **Why this matters more for an AI company OS than for normal SaaS**
>
> On May 1 2026, an AI coding agent (Cursor + Claude Opus 4.6) deleted a production company database — backups included — in nine seconds. The PocketOS incident is exactly what Galileo's customers will fear most. The v7 build discipline is not academic: it is the operational floor that prevents the same thing from happening to a Galileo tenant. Durable execution, snapshot-before-write, maker/checker, and pre-registered destructive-action gates are not features — they are the product.

# 3. Day-Zero Tenant Onboarding: How a Company Becomes a Galileo Tenant

A B2B customer's first hour with Galileo is not a setup wizard. It is an autonomous crawl. The Galileo Onboarding Crew — a special-purpose set of agents that runs once per tenant — connects to every system the company already uses, reads every document the company has, and builds the Company Brain before any department agent ships a single output. This section describes that flow end-to-end. It is the single most important customer-facing capability in the product, because it is what makes everything downstream possible.

## 3.1 The philosophy: onboarding is when the Brain is built

Tom Blomfield's argument — that the bottleneck for AI automation is not the models, it is the lack of structured executable knowledge inside companies — has a direct operational consequence. A Galileo tenant's first run cannot ask the operator 'tell me about your company.' The operator does not have time to do that, and even if they did, the answer they would give is incomplete. Galileo has to find out for itself by reading what the company already wrote down.

The mental model: pretend a brilliant new hire arrived this morning and had read-only access to every document, ticket, Slack channel, email thread, design file, and repo. By end of day, what would they know? That is the bar for what the Galileo Onboarding Crew produces in six hours, autonomously.

## 3.2 The Onboarding Crew (six agents, one workflow)

| **Agent** | **Role** | **Reads from** | **Produces** |
| --- | --- | --- | --- |
| Connector Agent | Authenticates every data source the operator approves | OAuth flows for GDrive, Slack, Gmail, Notion, GitHub, Linear, Microsoft 365, Dropbox, Confluence, Jira, Asana, ClickUp, HubSpot, Salesforce, Stripe, QuickBooks | Per-source credentials handed to downstream agents; Mirage workspace mount, if used, happens in-process inside Python agents. |
| Crawler Agent | Walks every mounted source enumerating documents | Credentials from Connector Agent; the Crawler imports `mirage-ai` for unified-filesystem access if it chooses that path, or uses discrete connector clients | Per-source manifest: list of documents, sizes, timestamps, hashes, content types |
| Ingestion Agent | Converts every document into clean markdown + extracted entities | Crawler output | Docling/MarkItDown markdown + insanely-fast-whisper transcripts + extracted tables/figures |
| Org-Mapper Agent | Synthesizes the org chart, departments, products, customers, vendors from the ingested corpus | Ingestion output + LLM reasoning | JSON org chart, department list, product catalog, customer list, vendor list, decision log |
| Skill-Selector Agent | Recommends an initial Skill set per department based on what was crawled | Org-Mapper output + Galileo Skill catalog | Per-department Skill manifest with version pins and rationale |
| QA Agent | Runs a checker pass against the synthesized artifacts and flags anything that smells off | All prior agent outputs | QA report with confidence scores per claim and an explicit list of items needing operator review |

The crew runs as a single Temporal workflow per tenant. Durable from the first connector authentication through final operator review. If a worker crashes mid-crawl, the workflow resumes from the last activity boundary; nothing is re-crawled unless the operator explicitly retriggers.

## 3.3 The eight-step onboarding flow

1. **Install.** Operator runs one of: (a) curl -fsSL https://galileoos.com/install.sh | bash for self-host; (b) git clone https://github.com/galileoos/galileo-os && make up for source install; (c) docker compose up -d from the repo for Stage 0 path; (d) one-click install on Coolify, Hetzner Cloud, or DigitalOcean Marketplace from Stage 1. All paths converge on the same Stage-0 docker-compose stack in Appendix B.
2. **Auth.** Operator signs in via Supabase Auth (Stage 0–1) or Authgear SSO (Stage 4 enterprise). First successful login provisions the tenant in Postgres with RLS isolation. No tenant data ever shares a row with another tenant.
3. **Connect data sources.** Web admin shows a connector picker (GDrive, Slack, Gmail, Notion, GitHub, Linear, …). Each OAuth grant is read-only by default; write scopes are not requested until the operator explicitly enables an action-taking department in Stage 1+. This is the destructive-action lockdown: a Galileo tenant cannot accidentally have its production database wiped on day one because Galileo never asked for write scope.
4. **Mount.** Connector Agent authenticates each source and writes the per-source credentials into Postgres `tenant_credentials` under AES-256-GCM with HKDF-SHA256 key derivation from the Stage 0 Ed25519 keypair (Stage 1+: Supabase Vault or Infisical per the Stage 0 closeout deferral, when production deployment lands — see [`docs/closeouts/CLOSEOUT_STAGE0.md`](../closeouts/CLOSEOUT_STAGE0.md) §4 row 12 trigger). **Per-source dispatch** is the live shape (PR-D, ADR-0005): MCP for sources with vendor-maintained MCP servers (`github` via `ghcr.io/github/github-mcp-server` Docker subprocess); direct SDKs for sources without (`slack_sdk`, `google-api-python-client`). Mirage's unified-filesystem abstraction remains available at Layer 5 for agent-side adopters that prefer one shell vocabulary across heterogeneous backends (see §4.6 and [`docs/decisions/0003-mirage-layer-relocation.md`](../decisions/0003-mirage-layer-relocation.md) for the forward path); per-source dispatch is the orthogonal pattern that Onboarding Crew agents use today. Agents that perform destructive operations record a pre-write snapshot artifact (Mirage's `workspace.snapshot()` for Mirage adopters; per-source backup otherwise); the Temporal workflow gates the destructive operation on that artifact's existence. The kernel does not snapshot.
5. **Crawl.** Crawler Agent walks every mount in parallel, producing a manifest. Caps are enforced per mount: no more than 50,000 documents per source on the first pass; no more than 6 hours total wall clock; no more than $50 in LLM and embedding spend (configurable). When a cap is hit, the workflow pauses and the operator is asked to expand the budget or scope.
6. **Ingest.** Ingestion Agent runs Docling on PDFs and Office files, MarkItDown on everything else, insanely-fast-whisper on audio (Slack huddle recordings, Gong calls, Zoom transcripts), Trafilatura on captured web pages, semchunk for chunking. Output: clean markdown + extracted entities per document, written into the Brain (pgvector + AGE + episodic events table).
7. **Synthesize.** Org-Mapper Agent reads the entire Brain and produces a structured org snapshot: who reports to whom, what departments exist, what products the company sells, who its top customers are, who its top vendors are, what its recent priorities have been (extracted from leadership communications). This is the artifact the operator reviews first — it is the proof that Galileo has actually understood the company.
8. **Calibrate Skills.** Skill-Selector Agent matches the org snapshot against the Galileo Skill catalog and produces a recommended Skill manifest per department. Example: a B2B SaaS company with a top-of-funnel content gap gets a different Marketing Skill set than a D2C brand with strong SEO but weak paid social. The recommendation is justified — every Skill pick comes with a one-sentence rationale citing what was found in the Brain.

## 3.4 The operator review (the one human gate)

After step 8, the workflow pauses and waits for operator approval. The operator sees three things in the web admin: the synthesized org snapshot, the recommended Skill manifest per department, and the QA Agent's list of anything that needs review. The operator can accept, edit, or reject. Nothing in the rest of the platform turns on until this approval signal is received.

This is the maker / checker separation expressed at the product level. The Onboarding Crew is the maker. The operator is the checker. The operator's approval is the only signal that promotes Galileo from 'read-only research mode' to 'active department agents enabled.'

## 3.5 Performance gate (pre-registered)

| **Dimension** | **Target** | **Failure mode** | **Mitigation** |
| --- | --- | --- | --- |
| Wall-clock time | <6 hours for 1M-document company on 4 vCPU VM | Crawl exceeds 6 hours | Auto-pause at 6h, operator asked to expand budget or sample. Failed pass produces CLOSEOUT_ONBOARDING_PERF.md. |
| LLM spend | <$50 per onboarding (1M docs) | Spend exceeds cap | Hard halt at LiteLLM gateway. Resume requires operator approval of new cap. |
| Org-snapshot accuracy | >90% of operator-reviewed claims marked accurate | Below 90% | Onboarding marked 'partial.' Operator does manual org definition. Failed pass triggers Org-Mapper Skill v2 spec. |
| Skill recommendation precision | >80% of recommended Skills kept by operator after review | Below 80% | Skill-Selector Agent's heuristics get re-calibrated. Calibration artifact committed before the agent is redeployed. |
| Destructive-action incidents | Zero across all onboardings, ever | One or more | Immediate platform-wide pause. Incident review per v7 closeout discipline. No retroactive softening of the destructive-action lockdown. |

## 3.6 Codebase onboarding (Stage 2+, opt-in)

For tenants whose engineering organization wants Galileo agents to have read access to source code — e.g., a Support agent that can answer 'what does our API actually do' by reading the API source instead of guessing — Stage 2 adds an opt-in codebase ingestion path. The Crawler (a Python agent) may use Mirage's `/github` mount to walk every repo the operator authorizes, or use the GitHub MCP server directly — choice is agent-local. Static analysis runs on each repo to extract a dependency graph, public function signatures, and README content. The result is a queryable code Brain alongside the document Brain.

Code ingestion is opt-in and read-only by default for the same reason document ingestion is: Galileo never takes write scope until it has earned it. A tenant that wants Galileo to file pull requests can do so explicitly in Stage 3, and even then every PR goes through a human review signal in the Temporal workflow before it merges.

## 3.7 Re-onboarding (continuous, not one-shot)

Onboarding is not a one-time event. Every connected source runs an incremental sync on a configurable schedule — new docs, new Slack threads, new Linear tickets get ingested as they arrive. The Brain is therefore a living, current substrate, not a frozen import. The autoresearch loop (described in Layer 4) runs nightly over the latest Brain state to keep the Skills current with how the company has actually evolved.

> **The Brain compounds**
>
> A Galileo tenant that has been running for six months has a Brain that no new competitor can replicate in less than six months. The defensible position is not the orchestration layer — that is commoditizing. The defensible position is the Brain, which gets richer every day that Galileo is connected. Day-zero onboarding is when the Brain is born; every subsequent day is when it compounds. This is the product.

# 4. The Stack, Layer by Layer

This section is the actual buy-list. Each layer specifies the primary choice, named alternatives, the language, what it replaces in a typical SaaS bundle, and why it was selected over the obvious competitors. Wherever a tool was discovered through Tom Doerr's MAGI//ARCHIVE feed (a curated stream of frontier OSS repos), that is noted.

## 4.1 Layer 0 — Identity, Tenancy, Foundation

Galileo OS is multi-tenant from the first commit. There is no 'we will retrofit multi-tenancy later' phase. This is the foundation that makes everything above it tenant-safe.

| **Function** | **Choice** | **Language** | **Replaces** | **Notes** |
| --- | --- | --- | --- | --- |
| Database + Auth + Storage + Realtime | Supabase | Go/Elixir/TS | Firebase + Auth0 (~$15K/year) | Postgres-native. Row-Level Security gives us tenant isolation for free. pgvector and AGE extensions give us vector + graph in one DB. ~73K stars, production at scale. |
| Enterprise SSO / OIDC | Authgear (or VoidAuth) | Go | Auth0 / Clerk / Okta | Supabase Auth handles end-user login. Authgear handles enterprise SSO when a Stage 4 customer needs SAML. Both surfaced via Tom Doerr's archive in May 2026. |
| Secrets management | Infisical | TS/Rust | HashiCorp Vault / Doppler | Self-hosted. Per-environment, per-project, per-folder. CLI + API + UI. Avoids the operational complexity of Vault. |
| Billing | Stripe Billing + LiteLLM cost meter | — | Recurly / Chargebee | Stripe Billing for subscription. LiteLLM emits per-tenant usage events that we pipe straight into Stripe metered billing. No homegrown meter. |
| Self-host installer | Coolify | PHP/TS | Heroku / Vercel / Render | Git-push-to-deploy. Auto SSL. 280+ one-click services. The 'Galileo on a single VM' experience is built on Coolify. |
| Production deploy | Kubernetes + Helm chart + ArgoCD/Kargo | — | Manual SSH / Capistrano | Kargo (recent OSS, in Tom Doerr's feed) handles GitOps lifecycle for Galileo's own SaaS. Customers can run the Helm chart directly. |
| Status / uptime | Uptime Kuma + status page | TS | Statuspage.io / BetterUptime | Free, self-hosted, beautiful. Public status page for the SaaS, internal dashboards for self-hosted customers. |

> **Why Supabase and not 'roll your own Postgres + auth'**
>
> Three reasons: Row-Level Security policies make tenant isolation a database-enforced invariant, not an application-layer convention you forget. The realtime engine gives the operator UI live ticket updates without us writing WebSocket plumbing. pgvector inside the same DB means the Company Brain (Layer 3) starts as a Postgres table — no separate vector DB until we hit hundreds of millions of vectors per tenant.

## 4.2 Layer 1 — LLM Gateway and Inference

Every LLM call in Galileo goes through a single proxy. This is non-negotiable. The proxy gives us provider failover, per-tenant cost caps, request logging, response caching, prompt versioning, and the ability to swap a tenant's model without touching agent code. Without a gateway, agents get hard-coded to one provider and the Paperclip 'no surprise bills' promise becomes impossible to deliver.

| **Function** | **Choice** | **Language** | **Replaces** | **Notes** |
| --- | --- | --- | --- | --- |
| Universal LLM proxy | LiteLLM | Python | OpenAI direct calls scattered through the codebase | Translates OpenAI format to 100+ providers (Claude, Gemini, Bedrock, Azure, Mistral, local). Per-key budgets. Fallback chains. Cost tracking emits to Prometheus. |
| Production local inference | vLLM | Python/CUDA | Self-hosted OpenAI | For customers who require on-prem (Stage 4 healthcare, defense). Continuous batching, paged attention, OpenAI-compatible. The standard for serious production deployment of open-weight models. |
| Dev / edge inference | Ollama | Go | — | For local dev environments and edge boxes. Go binary, one command to pull and run a model. Same OpenAI-compatible endpoint as vLLM, so dev/prod parity is preserved. |
| LLM trace observability | Opik | Python/TS | LangSmith / Langfuse paid tiers | Open-source. Captures prompt → response → tool calls → cost for every run. Integrates with LiteLLM, LangGraph, CrewAI, and the OpenClaw exporter Tom Doerr surfaced in April 2026. |
| Structured output | Instructor + Outlines | Python | Hand-written JSON parsers | Pydantic schemas in, validated objects out. Outlines forces grammar-constrained generation when we cannot tolerate a hallucinated key. |
| Prompt optimization | DSPy | Python | Manual prompt engineering | Stanford NLP. Programmatically optimizes prompts against a metric. Used in Stage 3 to auto-tune Skill prompts overnight. |
| LLM red-teaming | DeepTeam | Python | Manual jailbreak testing | Simulates 50+ vulnerability classes locally. Run against every Skill before we ship it. From Tom Doerr's May 6 2026 feed. |

> **Why LiteLLM and not a custom Go proxy**
>
> We considered writing a Go gateway. We rejected it. LiteLLM is one of the most actively maintained projects in the OSS LLM ecosystem, has every provider integration we will ever need already shipped, and is OpenAI-API-compatible — meaning every Python agent framework on earth talks to it natively. Building our own would be six months of catching up to feature parity for zero customer benefit. The right Go investment is one tier up, in the agent runtime and Temporal workers, where we have a defensible reason to own the code.

## 4.3 Layer 2 — Workflow and Agent Orchestration (the kernel)

This is the layer that earns Galileo the right to be trusted with real work. It is where 'an agent ran for 14 days, survived three worker restarts, two LLM provider outages, and one human approval delay, and finished correctly' becomes a routine occurrence rather than a heroic feat.

### Why durable execution matters for an agent OS

Most agent frameworks (LangChain, AutoGen, CrewAI in their default mode) keep state in process memory. If the worker dies mid-conversation, the agent loses everything. That is acceptable for a chatbot. It is unacceptable when the agent is mid-way through a customer refund, an outbound sales sequence, or a quarterly board update.

Temporal solves this by separating workflow definition (deterministic Go code) from activity execution (whatever you want, including LLM calls). Workflow state is persisted as an event history in Postgres or Cassandra; if the worker dies, a new worker reads the history and resumes exactly where the previous one left off. Twilio runs every message on Temporal. Coinbase runs every transaction on Temporal. Galileo runs every long-lived agent action on Temporal.

| **Function** | **Choice** | **Language** | **Replaces** | **Notes** |
| --- | --- | --- | --- | --- |
| Durable workflow engine | Temporal | Go | BullMQ + Redis + custom retry logic + Cron | Go-native server (matches your language preference). MIT-licensed. Production at Stripe, Netflix, Snap, Coinbase. Multi-region replication GA as of early 2026. |
| Inter-service / inter-agent bus | NATS JetStream | Go | Kafka / RabbitMQ | Go binary, sub-millisecond latency, true multi-tenancy via accounts (perfect fit for Galileo's tenant model). Runs on a Raspberry Pi if needed. Helm chart deploys in two commands. |
| Stateful agent graphs | LangGraph | Python | — | When an agent needs an explicit DAG with loops, conditionals, and human-in-the-loop checkpoints, LangGraph is the cleanest abstraction. Production-grade, used in serious AI systems. |
| Role-based agent crews | CrewAI | Python | — | When the metaphor is 'these four agents collaborate like a team' (e.g., the Marketing department), CrewAI is the right primitive. We do not pick LangGraph vs CrewAI vs Agno religiously — we pick the one that fits the workflow shape. |
| Fast simple agents | Agno | Python | — | When an agent is a single LLM + a few tools + memory, Agno is 10x lighter than LangChain. Used for the 'one agent, one job' department workers. |
| Agent runtime / executor | Custom Go binary 'agent-runner' | Go | — | Galileo's own. Wraps a Skill + an LLM + a tool set + a Temporal activity and runs it. The minimum viable agent. Inspired by the rsclaw 15MB Rust binary that Tom Doerr surfaced May 3, except we go Go for the rest of the ecosystem fit. |
| Multi-agent orchestrator UI | Custom — Galileo dashboard | TS | Paperclip (which we treat as inspiration) | Paperclip is brilliant as a reference design. We do not adopt it as-is because we need the org chart, ticket system, and budget controls to be tenant-scoped and tied to Temporal workflows. We rebuild the same UX on top of our kernel. |

### Concrete pattern: how a single agent action is structured

To make the abstraction concrete, here is exactly what happens when a Galileo customer asks the Marketing CEO agent to 'write the Q3 product launch email and post it to LinkedIn'.

```
// Go — workflow definition (Temporal)
// File: kernel/workflows/department_task.go
 
func DepartmentTaskWorkflow(ctx workflow.Context, in TaskInput) (TaskResult, error) {
    opts := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumAttempts:    5,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, opts)
 
    // 1. Load tenant Skills + budget guard
    var skills []SkillRef
    workflow.ExecuteActivity(ctx, LoadTenantSkills, in.TenantID, in.Department).Get(ctx, &skills)
 
    // 2. Run the agent (LangGraph, called as an activity over gRPC)
    var draft AgentOutput
    workflow.ExecuteActivity(ctx, RunAgent, RunSpec{
        TenantID: in.TenantID, SkillRefs: skills, Goal: in.Goal,
    }).Get(ctx, &draft)
 
    // 3. Human approval signal — workflow can sleep for days waiting
    var approved bool
    workflow.GetSignalChannel(ctx, "approval").Receive(ctx, &approved)
    if !approved { return TaskResult{Status: "rejected"}, nil }
 
    // 4. Execute the side-effecting tool calls (post to LinkedIn, send email)
    workflow.ExecuteActivity(ctx, CallMCPTool, draft.ToolCalls).Get(ctx, nil)
 
    // 5. Write the result back to the Company Brain
    workflow.ExecuteActivity(ctx, RecordOutcome, draft).Get(ctx, nil)
    return TaskResult{Status: "shipped"}, nil
}
```

Notice what is not in this code: no retry logic, no queue, no state machine, no resumable cursor, no cron. All of that is handled by Temporal. The workflow can sleep for two weeks waiting for human approval and resume the moment it arrives. The worker process can crash 100 times during the sleep and the workflow does not care. This is what 'durable execution' buys us.

## 4.4 Layer 3 — The Company Brain (Memory + Knowledge)

> **This is the moat**
>
> Tom Blomfield's argument in the source document is correct: the bottleneck for AI automation is no longer the models, it is the lack of structured, executable domain knowledge inside companies. Layer 3 is where Galileo turns a company's scattered knowledge — Slack threads, support tickets, old emails, Notion pages, recorded calls — into a queryable, citable, agent-actionable substrate. Every agent action above this layer is grounded in this layer.

### Three kinds of memory, one storage substrate

- **Semantic memory.** 'What does our refund policy say?' Embeddings over ingested documents and chats. Default: pgvector inside the same Postgres instance as Layer 0. Promoted to Weaviate when a tenant exceeds ~10M vectors.
- **Episodic memory.** 'Last Tuesday Jane in Support escalated a billing issue from customer ACME and resolved it by issuing a credit.' Time-stamped events, append-only, scoped to the tenant. Stored as Temporal workflow history plus a derived events table for fast querying.
- **Relational / graph memory.** 'Show me everyone connected to ACME — their SDR, the AE, the support tickets, the open invoices.' Postgres AGE extension at small scale; Neo4j when graph operations dominate. Avoids deploying a separate graph database in Stage 1.

| **Function** | **Choice** | **Language** | **Replaces** | **Notes** |
| --- | --- | --- | --- | --- |
| Vector DB (default) | pgvector inside Supabase Postgres | C | Pinecone ($) | One DB to back up, one DB to query, one DB to permission. RLS-scoped per tenant. Production-grade for the first ~10M vectors per tenant. |
| Vector DB (scale) | Weaviate | Go | Pinecone | Go-native (your preference), production at billions of objects, hybrid search out of the box. Promoted to per-tenant when a tenant outgrows pgvector. |
| Graph DB (default) | Apache AGE on Postgres | C | Neo4j | Same Postgres. openCypher queries. Avoids the operational burden of a separate graph DB until proven necessary. |
| Document conversion | Docling (IBM Research) + MarkItDown (Microsoft) | Python | Custom PDF parsers | Docling handles complex PDFs with tables, figures, formulas correctly. MarkItDown handles everything else (Word, Excel, PowerPoint, images, audio). One pipeline produces clean markdown for the Brain. |
| Web ingestion | Crawl4AI + Firecrawl | Python/TS | Custom scrapers | Crawl4AI for free, self-hosted scraping; Firecrawl when we need a managed crawl-the-whole-site endpoint. Both produce LLM-ready markdown. |
| Audio ingestion | insanely-fast-whisper | Python | Otter.ai / Rev | Whisper but 10–20x faster. Transcribes a 2-hour podcast in 2 minutes on consumer GPU. Used for sales call recordings, support calls, meetings. |
| Semantic chunking | semchunk | Python | Naive token-count split | Splits at natural boundaries instead of arbitrary token windows. Better chunks → better retrieval → better Skill performance. The unsung win in this stack. |
| Multimodal RAG | RAG-Anything (HKUDS) | Python | — | Handles text, tables, images, charts in one pipeline. Six lines to set up. Used in Stage 3 for design and proposal departments where the company's brand assets live. |
| Agent long-term memory | Memary / Mem0 pattern (custom Go service) | Go | — | Galileo's own service inspired by Memary, claude-mem, and Ori-Mnemos (all surfaced through Tom Doerr's archive). Wraps pgvector + AGE behind a clean 'remember/recall/forget' API that every agent uses. |
| Knowledge graph studio (admin UI) | WhyHow Knowledge Graph Studio | Python/TS | — | RAG-native KG editor. Used by tenant admins to inspect and curate their Brain. Shipped in Stage 3. |

> **Note on Mirage's placement.** Earlier drafts of this plan placed Mirage at Layer 3 as the unified data plane substrate, with a Stage 0 probe gate. The first plan-deviation in the project ([`docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md`](../closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md), 2026-05-13) relocated Mirage to Layer 5 once it became clear that Mirage's deployment model is in-process embedding inside Python or TypeScript agent code, with no Go SDK or native server. Mirage's value is preserved at the integrations layer (§4.6); Brain-state durability is covered by pgvector + AGE on the same Postgres instance as Layer 0. See [`docs/decisions/0003-mirage-layer-relocation.md`](../decisions/0003-mirage-layer-relocation.md) for the decision and reversal triggers.

### Ingestion pipeline (the moment a doc enters the Brain)

```
1. Source            → Drive sync, email forward, Slack export, web crawl, upload
2. Conversion        → Docling (PDF/Office) | MarkItDown (everything else) | Whisper (audio)
3. Cleanup + dedup   → Trafilatura (web) | Datatrove patterns (large)
4. Chunking          → semchunk (semantic boundaries, 200-800 tokens, 50 token overlap)
5. Enrichment        → entity extraction (spaCy + LLM), tag with department + confidentiality
6. Embedding         → bge-large-en-v1.5 OR text-embedding-3-large via LiteLLM
7. Storage           → pgvector (semantic) + AGE relations (graph) + events table (episodic)
8. Eval              → spot-check retrieval quality with Promptfoo eval set
 
Latency target: 30 seconds for a 50-page PDF on a single CPU core.
```

## 4.5 Layer 4 — Skills Library (the intelligence layer)

The previous infrastructure layers are commodity. Every serious AI company will build approximately the same kernel within a year. The Skills Library is what makes Galileo defensible. The thesis: a tenant's Skills are their actual operating procedures, encoded once and improving forever. The Skills are the asset; the agents are the runtime.

### The SKILL.md format as the standard

Galileo adopts Anthropic's SKILL.md format unchanged. This buys us instant compatibility with the broader Claude Code ecosystem, including the 175,000+ Skills already in agent-skills-cli, the gstack engineering Skills, gooseworks GTM Skills, the 86 PM Skills from RefoundAI's lenny-skills, and the email-marketing-bible from CosmoBlk — every one of which is OSS and discoverable through Tom Doerr's feed. Inventing a new format would lose all of that for cosmetic ownership.

On top of the standard format, Galileo adds three things the open ecosystem does not give us:

1. **Per-tenant private registry.** Each tenant has a private fork of the public Skill catalog plus an unlimited number of their own Skills. Skills are versioned (semver), signed, and gated by approval rules.
2. **Eval-on-every-version.** Every Skill ships with a Promptfoo eval suite. New version cannot deploy unless it beats the previous on the eval. No vibes, only metrics.
3. **Autoresearch loop.** The Karpathy autoresearch pattern from the source document is implemented as a Temporal workflow that runs nightly: generate samples, judge them with a stronger model, patch the Skill, re-evaluate, keep or discard. Twelve experiments per hour, ~100 per night, indefinitely.

| **Function** | **Choice** | **Language** | **Notes** |
| --- | --- | --- | --- |
| Skill format spec | Anthropic SKILL.md | Markdown | The standard. Frontmatter (name, description, license) + body (instructions). Optional bundled scripts and reference files. |
| Skill registry (per-tenant) | Custom Go service backed by Postgres | Go | CRUD, semver, sign, deploy, rollback. Webhooks on version change. ACLs by department. |
| Skill discovery (public) | Mirror of community catalogs | — | We mirror agent-skills-cli, paperclipai/companies, gstack, harness-100 (the 100-agent-team pack), and the curated indexes (Microck/ordinary-claude-skills with categorized search). |
| Skill eval | Promptfoo + custom judge models | Python/TS | YAML test files committed alongside each Skill. Run on every PR. Block merge on regression. |
| Skill validation | skill-check (TheDavidDias) | TS | Lints the SKILL.md format. Catches missing frontmatter, broken references, oversized prompts. Tom Doerr surfaced this May 4. |
| Autoresearch loop | Custom Temporal workflow inspired by karpathy/autoresearch | Python | Per-Skill, nightly. Generates candidate variants, scores against eval, retains improvements. Logged in the Brain. |
| Skill marketplace UI (Stage 3) | Custom (Next.js) | TS | Browse, install, fork, sell. ClipMart-style. Stripe Connect for revenue share with Skill authors. |

### Pre-built Skill packs to mirror at launch

- **paperclipai/companies** — 16 pre-built AI companies with 440+ specialized agents (April 2026, Tom Doerr feed). The single best starting point for Stage 3.
- **revfactory/harness-100** — 100 agent teams across 10 domains (May 4 2026 feed). Engineering-adjacent but huge cross-pollination value.
- **gooseworks/ai-goose-skills** — Growth and GTM tasks (April 27 2026). Direct fit for Marketing and Sales departments.
- **CosmoBlk/email-marketing-bible** — Comprehensive email marketing skill (April 29 2026). Plug into the Marketing department.
- **RefoundAI/lenny-skills** — 86 product management skills (April 29 2026). The PM department's starter pack.
- **FreedomIntelligence/OpenClaw-Medical-Skills** — 869 medical AI Skills (May 3 2026). One example of how rich the public Skill catalog is for vertical domains; reserved for any tenant that opts into the optional vertical Skill packs in Stage 4.
- **Karanjot786/agent-skills-cli** — 175,000+ agent Skills indexed and searchable. Used as the broad discovery surface.
- **Microck/ordinary-claude-skills** — Categorized library of hundreds of Claude Skills with search. Reference taxonomy.

> **Build vs adopt — the Skill question**
>
> Galileo does not write all of its Skills. It writes the contract (SKILL.md), the per-tenant registry, the eval harness, and the autoresearch loop. The Skill content itself is sourced from the open ecosystem, customized per tenant, and improved by the autoresearch loop. This is the same pattern that worked for npm, Docker Hub, and Hugging Face: own the registry, do not own the assets.

## 4.6 Layer 5 — Integrations and Tools

Agents need to do things in the world: send an email through Gmail, charge a card through Stripe, post to LinkedIn, query a CRM, click a button on a website. Layer 5 is the surface area of those side effects. Galileo deliberately uses three integration mechanisms in parallel because they suit different contexts.

| **Mechanism** | **Best for** | **Tool** | **Notes** |
| --- | --- | --- | --- |
| MCP servers | Programmatic, fine-grained tool calls direct from agents | Anthropic MCP + 500+ servers from awesome-mcp-servers | The standard. Each Galileo tenant gets a curated set of MCP servers gated by their plan and approval. |
| n8n workflows | Visual, multi-step integrations and triggers | n8n (self-hosted, MIT-fair-code) | 400+ pre-built integrations. Customers can wire their own without writing code. Replaces a $50K/year Zapier bill. |
| Browser automation | Anything the SaaS doesn't expose an API for | Playwright-MCP + browser-use-desktop | When a vendor has no API, Galileo agents drive a real browser. Playwright-MCP gives Claude a real browser with screenshots and click. browser-use-desktop ships a desktop app for human supervision. |
| Composio / Rube (managed catalog) | Customers who want one-click 'connect my Gmail' | Rube MCP server (already in your stack) + Composio | OAuth dance handled. Used as a fallback when self-hosting a specific MCP server is overkill. |

### Curated MCP server set for launch

From the awesome-mcp-servers catalog (27K stars, surfaced via the source document), the launch set covers every tool a typical small-to-mid business needs. Each is a separate process, sandboxed, with per-tenant credentials in Postgres `tenant_credentials` (Stage 0+) — promoted to Supabase Vault or Infisical when the Stage 0 closeout deferral §4 row 12 trigger fires.

**Per-source dispatch (PR-D, ADR-0005).** Not every "MCP server" entry in the table below ships as an MCP subprocess in the current Galileo. Where vendor-maintained MCP exists and is current (e.g., GitHub's `github/github-mcp-server`), Galileo invokes it via Docker subprocess. Where the previously-named reference MCP servers (`@modelcontextprotocol/server-*`) have been archived upstream and no vendor-maintained replacement exists (Slack, Google Drive), Galileo uses the vendor's direct Python SDK (`slack_sdk`, `google-api-python-client`). The table below names *which integration Galileo uses for each source*; whether that integration ships as MCP-subprocess or direct-SDK is dispatch-time decision per ADR-0005's four reversal triggers (Slack publishes vendor MCP; Google publishes vendor MCP; sixth source-kind lands; Docker becomes unavailable as a dev-stack prerequisite). See [`docs/decisions/0005-mcp-per-source-vs-mixed.md`](../decisions/0005-mcp-per-source-vs-mixed.md).

| **Department use** | **MCP servers** |
| --- | --- |
| Communication | Gmail, Outlook, Slack, Discord, Microsoft Teams, Telegram, WhatsApp Business API |
| Calendar | Google Calendar, Outlook Calendar, Cal.com (self-hosted scheduling) |
| CRM | HubSpot, Salesforce, Pipedrive, Twenty (OSS), NocoDB-as-CRM |
| Sales | Apollo, Lemlist, Instantly, Signal (OSS), FireEnrich (turn email lists into company datasets) |
| Support | Chatwoot, Frappe Helpdesk, Intercom, Zendesk, DocsGPT (OSS knowledge base) |
| Social | X/Twitter, LinkedIn, Instagram, TikTok, Reddit, Threads, Bluesky, brightbean-studio (self-hosted social manager) |
| Finance | Stripe, QuickBooks, Xero, Plaid, Invio (OSS invoicing), Cashew (OSS expense tracking) |
| Files / Docs | Google Drive, Dropbox, Notion, Confluence, S3, Supabase Storage |
| Web / Search | Web search (Brave/Bing/Tavily), Crawl4AI, Firecrawl, Playwright-MCP |
| Dev / Ops | GitHub, GitLab, Linear, Jira, Vercel, Supabase, mcp-toolbox for databases (Google) |
| E-commerce | Shopify, Shopware (OSS), Saleor (OSS), Storecraft (OSS, headless commerce + AI) |
| Vertical packs (Stage 4) | Domain-specific MCP servers — e.g., Epic FHIR / Cerner (healthcare), MLS / Zillow (real estate), Westlaw / Clio (legal), QuickBooks Advanced / NetSuite (financial services). Loaded only when a tenant opts into a vertical Skill pack. |
| Marketing analytics | Plausible (OSS), PostHog, GA4, Search Console, Ahrefs/Semrush |
| Unified data plane (agent-side, Python/TS) | `mirage-ai` (PyPI) and `@struktoai/mirage-*` (npm). Imported in-process by Python or TypeScript agents that want one shell vocabulary across heterogeneous backends (`grep -i 'refund' /slack/support/*.json` then `cat /github/api/refund.py`). Per-agent choice between Mirage and direct connector MCP servers; both coexist in the same Onboarding Crew. Mirage is **not** a kernel-side service — see §4.4 note on Mirage's placement and `docs/decisions/0003-mirage-layer-relocation.md`. |
| Notifications & alerting | Apprise (single API for 130+ notification services — Telegram, Slack, Discord, PagerDuty, email, SMS, webhooks) — surfaced via Tom Doerr archive May 10 2026 |
| Scraping (opt-in only) | Apify MCP server — opt-in per tenant for legitimate use cases only (own-brand monitoring, authorized competitive research, public-data analysis). Tenant must explicitly accept ToS and per-platform usage caps. NEVER enabled by default. |

> **Scraping is opt-in for legal and operational reasons**
>
> A B2B SaaS that ships scraping of LinkedIn, Instagram, TikTok, or X by default exposes its customers to ToS suits (LinkedIn v. hiQ precedent), platform bans, and an unbounded maintenance burden when scrapers break. Galileo's right answer is one well-isolated Apify MCP server, opt-in per tenant, with explicit ToS acknowledgment, per-platform rate caps, and audit logging of every request.
>
> Curated affiliate lists of scrapers (e.g., cporter202/social-media-scraping-apis) are not adopted as code. They are reference material for which Apify actors exist, evaluated case-by-case when a tenant has a legitimate need.

### Custom Galileo MCP servers (we write these)

Three MCP servers are not in the public catalog and we will write in Go. They are the ones that make Galileo Galileo, not 'a wrapper around someone else's stack.'

- **galileo-brain-mcp** — exposes the Company Brain (semantic + episodic + graph queries) to any agent. The single most-called tool in the system.
- **galileo-org-mcp** — exposes the org chart, roles, permissions, and budget caps. Lets agents introspect their own permissions before attempting a side-effecting tool call.
- **galileo-skill-mcp** — exposes the Skills library so agents can discover and load Skills at runtime, not just at boot.

## 4.7 Layer 6 — Department Modules

Each department is a curated collection of Skills + Tools + an org-chart slot, packaged as a Galileo 'module' that a tenant installs in one click. The intent is that on day one a Galileo customer can stand up a marketing department with a CMO, a content writer, a social manager, an ad ops agent, and a growth analyst — and have them coordinating before lunch.

The source document specifies eight department clusters and explicitly excludes engineering. We honor that. Galileo is the brain of the business; the engineers stay human.

| **Department** | **Default agents** | **Tools / OSS leverage** |
| --- | --- | --- |
| Marketing | CMO, content writer, social manager (per-channel), ad ops, growth analyst, SEO/AEO specialist | brightbean-studio (social), n8n flows, CosmoBlk email Skills, gooseworks GTM Skills, GetCito (AI/AEO/GEO monitoring), getCito + on-page-seo for audits, FireEnrich for lead data |
| Sales | AE, SDR, RevOps, deal coach, proposal writer | Signal (OSS sales intel), Cal.com (scheduling), Twenty CRM (OSS), Stripe links, FireEnrich, Apollo MCP |
| Customer Support | Tier-1 agent, escalation triage, knowledge curator, refund handler | Chatwoot OR Frappe Helpdesk (ticketing), DocsGPT (KB), MedRAX patterns for healthcare, ReceiptHero for billing disputes |
| Design | Brand guardian, asset producer, mockup builder, image generator | design.md format (Google Labs), Open Design (OSS Claude Design alternative), Snapframe, NPXSkillUI (reverse-engineers design systems into Skills) |
| Finance / Accounting | Bookkeeper, AR clerk, AP clerk, FP&A analyst, expense reviewer | Invio (invoicing), Cashew (expense), Financial Freedom (Mint OSS clone), QuickBooks/Xero MCP, Plaid for bank feeds, AI Bank Statement Document Automation pipeline (April 29 2026 Tom Doerr feed) |
| HR / People Ops | Recruiter sourcer, interview scheduler, onboarding owner, policy explainer, payroll liaison | OpenPostings (37k+ company ATS aggregator), ApplyPilot patterns for inbound resume screening, Cal.com, ERPNext HR module (Frappe), Notion as policy library |
| Project / Product Management | PM, sprint coordinator, customer feedback synthesizer, roadmap owner | Lenny Skills (RefoundAI, 86 PM Skills), prd-taskmaster (PRD generation), agent-kanban (agent-first kanban board, April 26 2026 feed) |
| Executive / Chief of Staff | CoS agent, board update writer, investor relations, calendar manager, exec-level summarizer | Custom Galileo Skills, Lenny Skills, GitHub of Tom's writings on systematized leadership, claude-memory-compiler for compiling chat logs into structured memos |

### Org chart model

Each tenant defines an org chart as data, not config. Roles inherit from a tenant's plan; agents are instantiated against roles; reporting structure determines which agents can ask which other agents for help. The same JSON-Schema document drives the operator UI, the access control engine, and the agent prompts (each agent is told who reports to it and who it reports to).

```
// Example org chart for a 12-person Galileo tenant
{
  "company": "Acme Coffee Co",
  "goal": "Grow MRR from $40K to $200K in 12 months",
  "roles": [
    { "id": "ceo",      "title": "CEO Agent",       "reports_to": null,    "budget_usd": 500 },
    { "id": "cmo",      "title": "CMO Agent",       "reports_to": "ceo",   "budget_usd": 300 },
    { "id": "content",  "title": "Content Writer",   "reports_to": "cmo",   "budget_usd": 80  },
    { "id": "social",   "title": "Social Manager",   "reports_to": "cmo",   "budget_usd": 60  },
    { "id": "ads",      "title": "Ad Ops",           "reports_to": "cmo",   "budget_usd": 60  },
    { "id": "sdr",      "title": "SDR Agent",        "reports_to": "ceo",   "budget_usd": 80  },
    { "id": "support1", "title": "Support Tier-1",    "reports_to": "ceo",   "budget_usd": 100 },
    { "id": "books",    "title": "Bookkeeper",       "reports_to": "ceo",   "budget_usd": 50  }
  ]
}
```

## 4.8 Layer 7 — Operator Interface

Three surfaces. Each is the right tool for one specific operating mode. We do not build one app and force every interaction through it.

| **Surface** | **Stack** | **Used for** |
| --- | --- | --- |
| Web admin | Next.js 16 + Tailwind + shadcn/ui + lucide-react + TanStack Query + Supabase JS | Primary control plane. Org chart, ticket queue, budget dashboard, Skill registry, Brain explorer, audit log. |
| Mobile companion | React Native + Expo + Supabase JS | Approval inbox, ticket review, budget alerts, agent on-call. Optimized for the founder traveling between San Francisco and Lagos who needs to approve sends from a phone. |
| Desktop shell | Tauri (Rust + WebView2/WKWebView) | For power users running Galileo against local-only LLMs (Ollama). Bundles vLLM/Ollama controls. ~10MB binary instead of an Electron 200MB blob — important when self-hosting on edge boxes. |
| Telegram bot | Go (telebot library) | The 'on the road' interface. The board-of-directors approval flow works over Telegram so founders never have to open the app. |
| WhatsApp gateway | open-bsp-api (self-hostable WhatsApp Business API, May 3 2026 feed) | Two-way ops over WhatsApp for tenants whose customers live there (large parts of Latin America, Africa, India, Southeast Asia). |
| Slack app | Slack Bolt (TS) | For tenants whose internal coordination already lives in Slack. Approve, query, and trigger department actions from a Slack channel. |

> **Why Tauri instead of Electron**
>
> Tauri ships ~10 MB binaries; Electron ships ~150 MB. Tauri uses the OS-native webview; Electron bundles a Chromium copy per app. Tauri's backend is Rust — the language you preferred and the right fit for a process that may run a local LLM alongside the UI. Every modern desktop AI tool that wants to feel native is on Tauri. We follow.

# 5. Language Strategy and Service Boundaries

You stated a preference for Go, Rust, C, Python, and React Native. The plan honors that preference while being honest about where each language is the right tool. The principle: pick the language for the workload, not for sentiment. The result is a polyglot but clean partition where every service has exactly one language and every language has a clearly bounded role.

| **Language** | **Owns** | **Specific services** | **Why** |
| --- | --- | --- | --- |
| Go | Orchestration kernel and all hot-path infrastructure | API gateway, agent-runner, Temporal workers, NATS subjects, custom MCP servers (galileo-brain-mcp, galileo-org-mcp, galileo-skill-mcp), webhook handlers, billing meter, Skill registry service | Concurrency model is the right abstraction for a multi-agent orchestrator. Static binaries deploy cleanly. Temporal is Go-native. NATS is Go-native. Matches your stated preference for the largest part of the codebase. |
| Python | Agent business logic and ML | LangGraph workflows, CrewAI crews, Agno agents, DSPy optimizers, Promptfoo eval harness, Docling/MarkItDown ingestion, autoresearch loop, embedding pipelines, eval judges | The LLM ecosystem is Python-first. Fighting that costs us months for no engineering benefit. Python services run as Temporal activities called over gRPC from the Go kernel — they are workers, not coordinators. |
| Rust | Edge, performance, native shells | Tauri desktop shell, optional 'edge runtime' (15MB binary inspired by rsclaw) for self-hosting on resource-constrained boxes, performance-critical Rust crates inside Go services via cgo where measured | Matches your stated preference. The Tauri shell matters because Galileo will be installed on developer laptops for self-hosting demos. Rust is also the right choice for any future trading or quant component (per Erio-Harrison/rust-trade pattern). |
| TypeScript | Web admin, Slack bot, custom n8n nodes | Next.js admin app, Slack Bolt app, custom n8n nodes for Galileo-specific triggers, lightweight Cloudflare Workers for public marketing site | TypeScript is the only sensible choice for the web admin given the React + Tailwind + shadcn ecosystem maturity. Resist the temptation to use it for anything backend. |
| React Native (Expo) | Mobile companion app | iOS + Android approval inbox, push notifications, biometric unlock, on-call view, budget alerts | Matches your stated preference. Expo's managed workflow gets us to TestFlight in a week. Shares the design system tokens with the web admin. |
| C | Reserved for embedded firmware (Stage 5+) | Out of scope for the first 18 months. Reserved for a possible Galileo edge appliance (the rack-mounted hardware version of the platform) and for any future hardware integration partners ship. | C is the right answer when there is silicon involved. There is no silicon in the first 18 months. |

## 5.1 Service boundary rules

1. A service is one language. No mixed-language services. If a Go service needs Python ML, the Python ML lives behind a gRPC endpoint, not as a subprocess.
2. All inter-service communication is gRPC over NATS, with protobuf contracts checked into a shared schemas repo. JSON is for external APIs only.
3. All long-lived agent actions are Temporal workflows. All short-lived tool calls are direct gRPC. The dividing line is 'could this run for more than 30 seconds across a process restart?' If yes, it's a workflow.
4. All LLM calls go through LiteLLM. No direct provider SDK usage outside the gateway. Enforced by lint rule and CI check.
5. All tenant-scoped data is read through a tenant context object that wraps a Postgres connection with the right RLS role set. No service ever queries Postgres without that context.

## 5.2 The repo layout

```
galileo/                                  # monorepo, Turborepo at the root
├── apps/
│   ├── web-admin/                        # Next.js 16, TS
│   ├── mobile/                           # React Native + Expo, TS
│   ├── desktop/                          # Tauri shell, Rust + TS
│   ├── slack/                            # Slack Bolt app, TS
│   └── tg-bot/                           # Telegram bot, Go
├── services/                             # backend services
│   ├── gateway/                          # Go — API gateway, tenant resolver
│   ├── agent-runner/                     # Go — runs agents as Temporal activities
│   ├── kernel/                           # Go — Temporal workflow definitions
│   ├── skill-registry/                   # Go — per-tenant Skill CRUD + signing
│   ├── brain-mcp/                        # Go — galileo-brain-mcp server
│   ├── org-mcp/                          # Go — galileo-org-mcp server
│   ├── skill-mcp/                        # Go — galileo-skill-mcp server
│   ├── ingestion/                        # Python — Docling, MarkItDown, Whisper pipelines
│   ├── eval/                             # Python — Promptfoo + DSPy + autoresearch
│   └── agents/                           # Python — LangGraph/CrewAI/Agno workers
├── packages/
│   ├── schemas/                          # protobuf, shared types
│   ├── ui/                               # shadcn-derived component library
│   ├── design-tokens/                    # generated from design.md
│   └── skill-fmt/                        # SKILL.md parser + validator (Go + TS bindings)
├── deploy/
│   ├── compose/                          # one-VM deployment (Coolify-friendly)
│   ├── helm/                             # Kubernetes deployment
│   └── terraform/                        # cloud infra (Hetzner / DO / AWS)
├── skills/                               # default Skill packs by department
│   ├── marketing/  sales/  support/  finance/  hr/  pm/  design/  exec/
└── docs/                                 # MkDocs Material
```

# 6. Phased Launch Roadmap

This is the most important section. Galileo will not be built and then launched. It will be launched five times. Each launch ships a complete product to a specific customer with a specific willingness to pay. The previous stage's customers stay; the next stage's product extends rather than replaces.

> **The principle this roadmap obeys**
>
> Every stage gate is defined as 'a paying customer has used this for a real outcome and stayed paying.' Internal milestones (services deployed, lines of code, infrastructure built) are explicitly not stage gates. This protects against the temptation to optimize for shipping infrastructure that nobody uses.

## 6.0 How Galileo gets installed (B2B distribution surface)

Galileo OS is software infrastructure for B2B businesses. It is installable in five ways. The same artifact runs in all of them — there is no 'cloud edition' that diverges from the OSS one. The distribution choice is purely a function of where the customer wants their data to live and how much operational work they want to do.

| **Path** | **Command** | **Audience** | **Available from** |
| --- | --- | --- | --- |
| One-line install (curl) | curl -fsSL https://galileoos.com/install.sh \| bash | SMB self-hosters, fast onboarding | Stage 1 |
| Git clone from GitHub | git clone https://github.com/galileoos/galileo-os && make up | Engineers who want to inspect the code first or fork | Stage 0 (private), Stage 1 (public) |
| Docker compose (single VM) | docker compose up -d (from the cloned repo) | Stage 0 demo, Stage 1 SMB single-box deployments | Stage 0 |
| Helm chart (Kubernetes) | helm install galileo galileo/galileo-os | Mid-market and enterprise with existing K8s | Stage 2 |
| One-click marketplace | Coolify / Hetzner / DigitalOcean / Vercel / Render templates | Operators who want zero infrastructure work | Stage 1 (Coolify), Stage 2 (others) |

All five paths converge on the same docker-compose service inventory documented in Appendix B. The Helm chart is the docker-compose translated to Kubernetes manifests with the same containers, the same env vars, and the same volume mounts. There is no 'enterprise version' of Galileo — there is one product, distributed multiple ways.

> **The dual GitHub + galileoos.com distribution model**
>
> Galileo is fully open-source. The canonical source-of-truth is https://github.com/galileoos/galileo-os. The galileoos.com website is the installer and documentation surface — it serves the curl|bash script, hosts the Helm chart repository, hosts release binaries, and runs the docs. A customer can install Galileo without ever visiting galileoos.com (just clone the repo), or without ever visiting GitHub (just run the curl command). Both surfaces pull from the same release artifacts.
>
> This dual surface is the standard pattern for OSS infrastructure products that need both developer credibility (GitHub) and operator convenience (a one-line installer). See: Coolify, Supabase, n8n, Temporal.

> **Stage 0  ·  Foundations  ·  Weeks 1–4**
>
> *Internal only. No paying customers. The kernel boots.*

### What ships at the end of Stage 0

- A monorepo with CI/CD, devcontainers, protobuf schemas, and the eight bare service skeletons.
- A self-hostable single-VM deployment via Coolify that runs: Postgres (Supabase), Temporal, NATS, LiteLLM, Opik, gateway, agent-runner.
- One demo agent — a 'Hello' agent that takes a prompt, runs it through LiteLLM, returns the response as a Temporal workflow result, and shows up in Opik with full trace.
- Cost meter wired end-to-end. Every LLM call is attributed to a tenant and a budget cap is enforced.
- Documentation: a 30-minute 'install Galileo on your laptop' walkthrough that any senior engineer can complete.
- Mirage layer-relocation closeout complete. Mirage is placed at Layer 5 (agent-side library, per-agent choice) rather than Layer 3 (kernel-side substrate). Decision documented in `docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md` and `docs/decisions/0003-mirage-layer-relocation.md`. The Workspace-interface verification harness from PR #10 is retained as a general kernel-side connector probe at `kernel/probe/connector/`.
- Onboarding Crew foundations. A scaffolded Connector Agent + Crawler Agent runs against an internal test workspace and produces a manifest. Full crew shipped in Stage 1.

#### Stage 0 gate

Three internal team members can spin up a fresh Galileo instance, register a tenant, set a $5 monthly budget, and run the demo agent 100 times without anyone looking at logs. The cost dashboard agrees with the Stripe metered-billing event count to the cent. The Mirage layer-relocation closeout, the canonical plan edits, and ADR-0003 exist and are committed.

#### Stage 0 critical risks

- Temporal operational complexity. Mitigation: use the official single-binary 'temporal server start-dev' for Stage 0; promote to Helm in Stage 2.
- LiteLLM proxy latency overhead. Mitigation: benchmark first, deploy LiteLLM as a sidecar to the agent-runner if hop adds >50ms.
- Mirage v0.0.x API instability. Mitigation: vendor a pinned commit and write integration tests; do not chase main. If breaking changes land, evaluate the patch cost vs. falling back to discrete MCPs at the Stage 1 entry gate.

> **Stage 1  ·  MVP — Galileo for Marketing  ·  Weeks 5–12**
>
> *Private beta with 10–25 design partners. First paying customers.*

### Why Marketing first

Three reasons. First, marketing has the clearest 'ship a deliverable' loop in any business — a post, an email, an ad — making it the easiest department to wire end-to-end. Second, marketing budgets are the most flexible in any company; founders are willing to experiment. Third, the source document calls out marketing as one of the explicit department clusters and the supplied OSS surface (gooseworks Skills, CosmoBlk email-marketing-bible, brightbean-studio, FireEnrich, gstack growth marketers) is the deepest of any department.

### What ships at the end of Stage 1

- Five default agents: CMO (planner/coordinator), Content Writer, Social Manager, Ad Ops, Growth Analyst.
- Onboarding Crew GA. The full six-agent crew from §3 runs autonomously on every new tenant. The Stage 1 gate explicitly tests onboarding-crew performance against the §3.5 pre-registered targets — a tenant cannot be counted toward the gate if their onboarding failed the perf gate.
- Web admin: org-chart UI (populated by Org-Mapper Agent from onboarding), ticket queue, calendar of scheduled posts, budget dashboard, Brain explorer with per-document provenance.
- Brain: tenant connects sources via the Connector Agent (GDrive, Slack, Gmail, X, LinkedIn, GA, Plausible, Stripe). The Brain is built in the first hour, not by manual upload.
- Connectors: read-only by default for every source. Write scope requested explicitly and per-action when a department goes live.
- React Native mobile app: approval inbox + push notifications. The 'approve before this goes out at 9am' loop.
- Telegram bot: same approval loop, fallback for travel.
- Skills: ~40 marketing Skills mirrored from gooseworks + CosmoBlk + custom Galileo Skills written for the design partners.
- Eval: Promptfoo runs on every Skill PR; agents have measurable quality scores per Skill version.
- Destructive-action lockdown enforced platform-wide. Every DELETE / DROP / rm -rf / send-email / post-publish action takes a Temporal signal from a human or a pre-authorized policy before executing. Snapshot-before-write via Mirage on tenants where Mirage probe passed.

#### Stage 1 gate

Five paying customers (~$200–$500/month each) using Galileo to ship at least one marketing artifact per day for two consecutive weeks, with a measured reduction in their existing tooling spend (Buffer, Hootsuite, Jasper, Copy.ai, etc.) of at least 50%.

#### Stage 1 pricing model (initial)

- Tier 1 'Starter': $99/month + LLM pass-through. One department, two agents, $50/month included LLM credit.
- Tier 2 'Team': $399/month + LLM pass-through. One department, full agent set, $200/month included LLM credit, 1 user seat.
- Tier 3 'Studio': $1,499/month + LLM pass-through. Two departments, full agent set, $1,000/month included credit, 5 user seats, custom Skills support.

> **Stage 2  ·  AI Office — Multi-Department Beta  ·  Weeks 13–24**
>
> *Public beta. Open waitlist. The platform becomes useful end-to-end.*

### What ships at the end of Stage 2

- Three more departments: Sales, Customer Support, Design. Each shipped with default agents, default Skills, and default tool wiring.
- Inter-agent messaging fully on NATS. The Sales SDR can hand off to the Marketing CMO without going through human ops.
- MCP integration layer with at least 30 connectors live (the curated set from §4.6).
- Skill marketplace UI v1: browse, install, version, fork. No payments yet. ClipMart inspiration: 16 paperclipai/companies templates available as one-click installs.
- Multi-company support: a single user account can administer multiple Galileo tenants (a consultancy with five client deployments, for example).
- Telegram bot becomes a first-class Slack-equivalent surface, not just an alert channel. Full ticket review and approval over chat.
- WhatsApp Business gateway live for tenants whose customers live in WhatsApp-first markets.
- Promoted infrastructure: Helm chart for Kubernetes, multi-region Postgres, NATS clustering. Single-VM still supported as the on-ramp.

#### Stage 2 gate

50 paying customers, $25K MRR, NPS ≥ 30 from the design partner cohort. At least three customers running two or more departments simultaneously with documented inter-departmental handoffs.

#### Stage 2 release sequence inside the 12 weeks

1. Weeks 13–14: Sales department alpha (internal).
2. Weeks 15–16: Sales GA + Support department alpha.
3. Weeks 17–18: Support GA + Design department alpha.
4. Weeks 19–20: Design GA + connector expansion to 30.
5. Weeks 21–22: Skill marketplace UI + multi-company.
6. Weeks 23–24: Public beta launch event + Product Hunt + Show HN.

> **Stage 3  ·  Galileo OS 1.0 GA  ·  Months 7–12**
>
> *General availability. Self-serve sign-up. Sales-led for the upper tier.*

### What ships at the end of Stage 3

- All eight department clusters from the source document live: Marketing, Sales, Support, Design, Finance/Accounting, HR/People Ops, Project/Product Management, Executive/Chief of Staff. Engineering remains explicitly excluded.
- Skill marketplace v2: payments, revenue share with Skill authors via Stripe Connect, ratings, eval scores visible on every Skill page.
- Template companies: install a fully populated tenant in one click — 'Galileo for a Real Estate Agency,' 'Galileo for an E-commerce Brand,' 'Galileo for a Content Studio,' 'Galileo for a Telehealth Practice.' We mirror and extend paperclipai/companies.
- Custom Skill builder IDE in the web admin. Non-technical operators can create Skills via a guided wizard; senior operators can drop into raw markdown.
- Autoresearch loop running per-tenant nightly. Skill quality scores improve measurably week over week without operator intervention.
- White-label option: agencies can resell Galileo under their own brand for an upcharge.
- Self-host installer: 'curl | bash' lands a complete Galileo on a fresh Hetzner box in under 10 minutes via Coolify.
- Stripe Billing fully wired with usage-based metered LLM markup, included credits, overage handling, and clear tenant-facing billing dashboards.

#### Stage 3 gate

$200K MRR. 500 paying customers. At least 50 of them on the self-host plan. The Skill marketplace has at least 25 third-party authors with at least one paid Skill installed by another tenant.

#### Stage 3 expansion bets (run in parallel)

- Galileo Cloud — managed multi-tenant hosting for tenants who do not want to self-host. Higher margin tier.
- Galileo Pro Services — a small forward-deployed engineering team that ships custom Skills and integrations for high-touch tenants. Mirrors the YC AI-native services thesis from the source document.
- Open-core dual license: core is OSS, advanced features (SSO, audit log retention >90 days, multi-region replication, dedicated single-tenant) are paid.

> **Stage 4  ·  Enterprise + Vertical Editions  ·  Year 2**
>
> *Compliance-grade deployment. Industry-specific editions. The Stripe Atlas of business AI.*

### What ships during year 2

- Compliance: SOC 2 Type II completed; ISO 27001 in flight; HIPAA-ready edition signed off by an outside auditor.
- Single-tenant dedicated deployments for enterprise. Air-gapped install option for defense / financial services.
- Multi-region active-active via Temporal Multi-Region Replication (GA in 2026).
- Galileo Healthcare — HIPAA edition. Pre-bundled with Epic FHIR, Cerner, and a curated medical Skills pack (e.g., FreedomIntelligence's 869-Skill catalog).
- Galileo Legal — pre-bundled with Westlaw/LexisNexis MCP, contract review Skills, deposition prep agents.
- Galileo Real Estate — pre-bundled with MLS, lead scoring agents, listing posters, transaction coordinators.
- Galileo Financial Services — pre-bundled with NetSuite, advanced QuickBooks, Plaid, and SOC-2-attested audit trails.
- Per-tenant fine-tuned models: LLaMA-Factory + Unsloth pipelines that can spin up a tenant-specific Llama or Qwen variant trained on the tenant's Brain.
- Audit log archival to S3 / Glacier with cryptographic integrity proofs (Merkle tree per day).
- Procurement-friendly motion: SOC 2 reports auto-shared via Vanta / Drata, MSA templates, security questionnaire library.

#### Stage 4 gate

$2M ARR. At least 10 enterprise tenants on the dedicated tier. At least one vertical edition shipped with five or more named-brand customers in production.

## 6.1 Rollup timeline view

|  | **Stage 0** | **Stage 1** | **Stage 2** | **Stage 3** | **Stage 4** |
| --- | --- | --- | --- | --- | --- |
| Window | Wk 1–4 | Wk 5–12 | Wk 13–24 | M 7–12 | Year 2 |
| Customers | 0 | 5–25 | 50+ | 500+ | 10+ enterprise; 1000s SMB |
| Departments | 0 (demo only) | 1 (Marketing) | 4 (+ Sales, Support, Design) | 8 (all) | 8 + verticals |
| Connectors | 0 | 8 | 30 | 100+ | 100+ + vertical-specific |
| Skills (default) | 0 | ~40 | ~150 | ~500 | Marketplace + verticals |
| Hosting model | Local dev | Single VM | Single VM + cloud | Helm + cloud | Helm + dedicated + air-gap |
| Compliance | — | — | Baseline security review | SOC 2 Type I in flight | SOC 2 II + HIPAA + ISO 27001 |
| Headcount target | 1–3 | 3–5 | 5–10 | 10–20 | 20–40 |
| MRR / ARR target | $0 | $5K MRR | $25K MRR | $200K MRR | $2M ARR |

> **What this roadmap explicitly does NOT do**
>
> It does not try to ship all 8 departments at once. The source document's 'AI agency' frameworks (gstack, Paperclip, harness-100) are exciting precisely because they enumerate everything — but a launch needs focus, and Marketing-then-Sales-then-Support is the right sequencing because each downstream department compounds the value of the previous one.
>
> It does not try to win the engineering-agent market. Galileo competes with ServiceNow, HubSpot, and the SaaS bundle — not with Cursor, Claude Code, OpenHands, Aider, or Cline. We adopt those tools as customers (we use them to build Galileo); we do not build a competing product.
>
> It does not try to be a coding agent platform. The user is explicit about excluding engineering. We honor that and gain a clearer positioning for it.

# 7. Complete Bill of Materials

Every external technology in the Galileo stack, organized by layer, with license, primary language, what it replaces, and the stage at which it enters the build. This is the canonical reference — when an architectural question arises, this table answers it.

| **Tool** | **Layer** | **Lang** | **License** | **Replaces / why** | **Stage** |
| --- | --- | --- | --- | --- | --- |
| Supabase | L0 | Multi | Apache-2.0 | Firebase + Auth0 (~$15K/yr); single Postgres for DB+auth+storage+vector+realtime | 0 |
| Authgear | L0 | Go | Apache-2.0 | Auth0/Clerk for enterprise SSO/SAML/OIDC | 4 |
| Infisical | L0 | TS/Rust | MIT | HashiCorp Vault; per-env per-project secret rotation | 0 |
| Stripe Billing | L0 | — | Commercial | Recurly/Chargebee; metered billing wired to LiteLLM events | 1 |
| Coolify | L0 | PHP/TS | Apache-2.0 | Heroku/Vercel for self-host installer | 0 |
| Helm + Argo/Kargo | L0 | Go | Apache-2.0 | Manual K8s deploys; GitOps lifecycle | 2 |
| Uptime Kuma | L0 | TS | MIT | Statuspage.io / BetterUptime | 1 |
| LiteLLM | L1 | Python | MIT | Direct OpenAI calls; provider failover and cost meter | 0 |
| vLLM | L1 | Py/CUDA | Apache-2.0 | Self-hosted OpenAI for production local inference | 3 |
| Ollama | L1 | Go | MIT | Dev/edge local inference; same OpenAI shape as vLLM | 0 |
| Opik | L1 | Py/TS | Apache-2.0 | LangSmith / Langfuse paid; full LLM trace observability | 0 |
| Instructor | L1 | Python | MIT | Hand-written JSON parsers; pydantic-typed structured outputs | 1 |
| Outlines | L1 | Python | Apache-2.0 | Manual prompt engineering for structured outputs; grammar-constrained generation | 1 |
| DSPy | L1 | Python | MIT | Manual prompt tuning; programmatic prompt optimization | 3 |
| DeepTeam | L1 | Python | Apache-2.0 | Manual jailbreak testing; LLM red-teaming | 2 |
| Temporal | L2 | Go | MIT | BullMQ + Redis + cron + custom retry; durable workflows | 0 |
| NATS JetStream | L2 | Go | Apache-2.0 | Kafka/RabbitMQ; multi-tenant pub/sub bus | 0 |
| LangGraph | L2 | Python | MIT | —; stateful agent graphs | 1 |
| CrewAI | L2 | Python | MIT | —; role-based agent teams | 1 |
| Agno | L2 | Python | MPL-2.0 | —; fast simple agents | 1 |
| agent-runner (custom) | L2 | Go | AGPL+Commercial | —; Galileo's own minimal agent runtime | 0 |
| pgvector | L3 | C | PostgreSQL | Pinecone (default vector DB) | 0 |
| Apache AGE | L3 | C | Apache-2.0 | Neo4j (default graph DB on Postgres) | 1 |
| Weaviate | L3 | Go | BSD-3 | Pinecone at scale; promoted when tenant exceeds 10M vectors | 3 |
| Docling | L3 | Python | MIT | Custom PDF parsers; IBM's high-fidelity converter | 0 |
| MarkItDown | L3 | Python | MIT | Custom Office parsers; Microsoft's universal-to-markdown | 0 |
| Crawl4AI | L3 | Python | Apache-2.0 | Custom scrapers; LLM-ready web crawl | 1 |
| Firecrawl | L3 | TS | AGPL | Crawl4AI for managed/full-site crawls | 2 |
| insanely-fast-whisper | L3 | Python | MIT | Otter.ai / Rev; audio ingestion | 2 |
| semchunk | L3 | Python | MIT | Naive token split; semantic chunk boundaries | 0 |
| RAG-Anything (HKUDS) | L3 | Python | MIT | —; multimodal RAG | 3 |
| trafilatura | L3 | Python | Apache-2.0 | BeautifulSoup; clean web extraction | 1 |
| Memary pattern (custom) | L3 | Go | AGPL+Commercial | —; agent long-term memory service | 1 |
| WhyHow KG Studio | L3 | Py/TS | MIT | —; admin UI for the Brain's graph | 3 |
| Anthropic SKILL.md | L4 | Markdown | — | —; the format itself | 0 |
| Promptfoo | L4 | TS | MIT | Manual eval; Skill quality regression tests | 1 |
| skill-check | L4 | TS | MIT | Hand validation; lints SKILL.md format | 1 |
| paperclipai/companies | L4 | Markdown | MIT | —; 16 pre-built companies, 440+ agents | 3 |
| revfactory/harness-100 | L4 | Markdown | MIT | —; 100 agent teams across 10 domains | 3 |
| RefoundAI/lenny-skills | L4 | Markdown | MIT | —; 86 PM Skills | 2 |
| gooseworks/ai-goose-skills | L4 | Markdown | MIT | —; growth and GTM Skills | 1 |
| CosmoBlk/email-marketing-bible | L4 | Markdown | MIT | —; comprehensive email marketing Skills | 1 |
| FreedomIntelligence Medical | L4 | Markdown | MIT | —; 869 medical AI Skills (vertical pack option) | 4 |
| Karanjot786/agent-skills-cli | L4 | Multi | MIT | —; 175,000+ Skill discovery surface | 3 |
| MCP (anthropic spec) | L5 | Multi | MIT | Custom REST plumbing; standard tool protocol | 0 |
| awesome-mcp-servers | L5 | Multi | MIT | Custom integrations; 500+ server catalog | 1 |
| n8n | L5 | TS | Fair-code | Zapier ($50K/yr); visual workflow + 400+ apps | 1 |
| Composio / Rube MCP | L5 | Multi | Mixed | Per-app OAuth plumbing; managed catalog of connectors | 2 |
| Playwright-MCP | L5 | TS | Apache-2.0 | —; browser-as-tool for agents | 2 |
| browser-use-desktop | L5 | Py/TS | MIT | —; supervised browser automation | 2 |
| Cal.com | L5 | TS | AGPL | Calendly; self-hosted scheduling | 2 |
| brightbean-studio | L6 | TS | MIT | Buffer/Hootsuite; self-host social management | 1 |
| Signal (jay-sahnan) | L6 | TS | MIT | Apollo/Outreach; OSS sales intel + outreach | 2 |
| Chatwoot | L6 | Ruby | MIT | Intercom/Zendesk; OSS customer support inbox | 2 |
| Frappe Helpdesk | L6 | Py/JS | AGPL | Zendesk for ERPNext customers | 2 |
| DocsGPT | L6 | Python | MIT | Intercom Articles; OSS knowledge base + chat | 2 |
| Twenty CRM | L6 | TS | AGPL | Salesforce/HubSpot; OSS modern CRM | 2 |
| NocoDB | L6 | TS | AGPL | Airtable; flexible CRM/ops backend | 2 |
| Invio | L6 | TS | MIT | FreshBooks/Bill.com; OSS invoicing | 3 |
| Cashew | L6 | Dart | MIT | Mint/YNAB; OSS expense tracking | 3 |
| FireEnrich | L6 | TS | MIT | Apollo/Clearbit; turn email lists into company data | 2 |
| ERPNext modules (Frappe) | L6 | Py/JS | GPL | NetSuite; HR, payroll, inventory | 3 |
| Open Design / DESIGN.md | L6 | Multi | MIT | Figma plugins; design-system-as-Skill | 2 |
| agent-kanban | L6 | TS | MIT | Linear/Jira; agent-first kanban | 3 |
| Next.js 16 | L7 | TS | MIT | Custom SPA; web admin | 0 |
| Tailwind + shadcn/ui | L7 | TS | MIT | Bootstrap; design system | 0 |
| React Native + Expo | L7 | TS | MIT | Native iOS/Android; mobile companion | 1 |
| Tauri | L7 | Rust | MIT/Apache | Electron; desktop shell | 2 |
| Slack Bolt | L7 | TS | MIT | Custom Slack integration | 2 |
| open-bsp-api | L7 | Multi | AGPL | Twilio WhatsApp; self-host WhatsApp Business gateway | 2 |
| Prometheus + Grafana | X | Go | Apache-2.0 | Datadog ($); infra metrics | 0 |
| PostHog | X | Py/TS | MIT | Mixpanel + Sentry replay; product analytics | 1 |
| Plausible | X | Elixir | AGPL | Google Analytics; privacy-respecting web analytics | 1 |
| Keep | X | Python | MIT | PagerDuty; AIOps + alert mgmt | 3 |
| Sentry (self-host) | X | Py/TS | FSL | Sentry SaaS; error tracking | 1 |
| Apprise | X | Python | BSD-2 | —; single API for 130+ notification services (Tom Doerr May 10 2026) | 1 |
| Vanta / Drata | X | — | Commercial | Manual SOC 2 compliance | 4 |
| Mirage (strukto-ai) | L5 (agent-side) | Py/TS | Apache-2.0 | Agent-side unified VFS option, per-agent choice alongside discrete MCP wrappers | 0 closeout / 1 agent-side adoption |
| Apify MCP (opt-in) | L5 | JS/TS | Mixed | —; managed scraping for legitimate use cases only | 2 |

# 8. Risk Register

Every architecturally consequential decision in this plan carries risk. Naming the risks early — and pairing each with a concrete mitigation, an owner, and a stage at which it bites — is how the team avoids stage-shipping a foreseeable failure. The table below is the canonical risk list for the first eighteen months. It is reviewed at every stage gate.

| **Risk** | **Severity** | **Stage it bites** | **Mitigation** | **Owner** |
| --- | --- | --- | --- | --- |
| LLM provider outage or pricing shock | High | 1+ | All calls go through LiteLLM with multi-provider failover. Provider quotas reviewed quarterly. Stage 3 adds vLLM self-hosting to break the dependency. | CTO |
| Temporal operational complexity | Medium | 0+ | Use Temporal Cloud through Stage 2, self-host only after on-call team has run a Temporal cluster in staging for 60+ days. Document runbooks before promotion. | Platform |
| Multi-tenant data leak (cross-tenant) | Critical | 1+ | Postgres RLS enforced at the role level, never the application. Tenant context object is the only path to a DB connection. Quarterly third-party pen test from Stage 2. | Security |
| Skill quality drift (silent regression) | High | 1+ | Promptfoo eval runs on every Skill version. Skill cannot deploy if eval score drops. Nightly autoresearch loop benchmarks the top-100 Skills against a held-out judge set. | Skills |
| Agent cost runaway (loop, hallucinated tool calls) | High | 1+ | Per-tenant monthly cap enforced at LiteLLM. Per-workflow soft cap enforced by agent-runner. Agents that hit the cap halt and post a Brain note for human review. | Platform |
| OSS license collision (AGPL viral effects) | Medium | 2+ | AGPL components (Chatwoot, Twenty CRM, Cal.com, Frappe Helpdesk, Firecrawl) are run as external services with documented network boundaries. No AGPL code is statically linked into the Galileo core. Legal review at Stage 2. | Legal |
| Build-vs-buy on Skill content (we underbuild) | Medium | 1-3 | Mirror paperclipai/companies, harness-100, lenny-skills, gooseworks, CosmoBlk, Medical-Skills as the seed catalog. Hire one Skills curator at Stage 2. Never write Skills from scratch when a public one is good enough. | Skills |
| Regulatory exposure (Stage 4 vertical editions) | Critical | 4 | Compliance-grade verticals (Healthcare, Legal, Financial Services) run single-tenant only, in dedicated VPCs, with separate keys. HIPAA / SOC 2 audits are vertical-scoped, not horizontal. Compliance review with outside counsel before launch of each vertical edition. | CEO |
| Marketplace race to bottom (low-quality Skills) | Medium | 3+ | Marketplace Skills must pass eval gate plus human review before listing. Revenue share favors high-rated Skills. Refund policy on Skill purchases up to 30 days. | Marketplace |
| MCP standard fragments (vendor extensions) | Low | 2+ | Stick to the spec'd MCP surface. Vendor-specific extras are wrapped in adapters. We test against the official MCP conformance suite quarterly. | Platform |
| Self-host installer breaks across distros | Medium | 0-3 | Coolify is the supported installer. Helm chart is the supported K8s path. Anything else is community-supported. Runtime tested on Ubuntu 24.04 and Debian 12 in CI. | DevRel |
| Competitor (Salesforce Agentforce, ServiceNow Now Assist) | Medium | 2+ | Win on self-host, transparent pricing, OSS-first. Don't try to outspend incumbents on enterprise sales. Target SMB and mid-market through Stage 3. | CEO |
| Long context window costs blow up the Brain | Medium | 1+ | Aggressive RAG over Brain queries. No naive 'stuff full Brain into context' calls. Per-query token budget enforced in the brain-mcp server. | Brain |
| Destructive agent action (Cursor/PocketOS scenario) | Critical | 1+ | Three structural defenses: (1) Read-only OAuth scopes for every Onboarding Crew connector; write scope requested per-action only after operator approval. (2) Every destructive call (DELETE, DROP, rm, send, post, publish, transfer) goes through a Temporal signal channel — workflow halts and waits for human approval signal before executing. (3) Pre-write snapshot artifact required by the Temporal gate. The agent that performs the destructive operation provides a snapshot artifact reference before the call — a content-addressable hash, a storage URL, or a backup-system-issued token that the Temporal workflow validates exists and is reachable. Valid artifacts come from either source-native mechanisms (S3 versioning, Postgres logical backup, GDrive revision history, Mirage `workspace.snapshot()` for Python agents using Mirage) or Galileo-produced backups (full read-and-store of the affected state before the destructive call). The kernel enforces that an artifact exists and is reachable; the kernel does not produce the artifact itself. Brain-state durability (Postgres PITR) is a separate defense covering catastrophic recovery, not per-operation rollback. No exceptions, no overrides via prompt injection. | Security |
| Mirage v0.0.x API breakage | Medium | 0–2 | Risk scope is agent-side, not platform-wide. Python agents pinning `mirage-ai==<version>` should update deliberately; integration tests in `agents/*/` against the pinned version run on every Galileo CI build. If Mirage ships breaking changes that don't justify the patch cost, affected agents fall back to discrete connector clients. The kernel is unaffected. | Platform |
| Onboarding Crew misses perf gate | High | 1 | Pre-registered §3.5 gate. If a Stage 1 tenant's onboarding fails the gate, that tenant does not count toward the Stage 1 gate, and a CLOSEOUT_ONBOARDING_TENANT_X.md is produced. Repeat failures trigger an Onboarding Crew v2 spec, not relaxed gates. | Brain |
| Scraping ToS violations (LinkedIn v. hiQ scenario) | High | 2+ | Apify MCP is opt-in only with explicit per-tenant ToS acknowledgment. Default install does not include it. No social-media scraping is shipped as a default Marketing Skill. Legal review of any scraping use case before activation. | Legal |

> **The three non-negotiable risks**
>
> Cross-tenant data leakage, Skill quality drift, and destructive agent actions are the three risks that, if they materialize, end the company. Everything else is recoverable. The architecture pays a real cost for all three — Postgres RLS at every layer, eval-on-every-Skill-version, and the destructive-action lockdown — and that cost is not negotiable down to ship faster.
>
> The PocketOS incident on May 1 2026 — a coding agent deleting an entire production database in nine seconds, backups included — is the canonical failure mode for this category of product. Galileo's job is to be the product that this cannot happen on. Not because we wrote a careful prompt; because the architecture makes it structurally impossible.

# 9. Build vs Buy vs Fork — Explicit Decisions

For every major component there is a question of whether to build it, adopt it as-is, or fork it. The default is adopt. The exception is anything that is part of the durable moat (the Brain, the agent kernel, the Skill registry) — those we build, because they are the things customers cannot get anywhere else. Everything else is a commodity, and a startup wins by not rewriting commodities.

| **Component** | **Decision** | **Reasoning** |
| --- | --- | --- |
| Durable workflow engine (Temporal) | Adopt | Battle-tested at Stripe / Snap / Coinbase scale. Multi-region GA in 2026. Go-native. We could not justify rebuilding this in years. Use Temporal Cloud through Stage 2, then self-host. |
| Message bus (NATS JetStream) | Adopt | Go-native, sub-millisecond, multi-tenant. Already supports the patterns we need (subjects, JetStream durability, key-value store). Kafka would be over-spec for our scale. |
| LLM gateway (LiteLLM) | Adopt | Python is fine here — the gateway is not on the hot path of agent reasoning, and LiteLLM's provider coverage is the deepest in OSS. Rebuilding it in Go would consume two engineer-quarters for marginal gain. Wrap with our own thin Go shim only for tenant context and budget enforcement. |
| Agent frameworks (LangGraph + CrewAI + Agno) | Adopt all three | Different agents have different needs. LangGraph for stateful long-running flows, CrewAI for role-based teams, Agno for fast simple agents. We do not pick one — we expose all three through the agent-runner. |
| Agent runtime / runner | Build (Go) | This is the kernel. It enforces tenant context, budget caps, Skill loading, tool authorization. Every other vendor's runtime mixes business logic with execution; ours separates them. Forty hours of Go for the first version. |
| Company Brain | Build (Go) on top of Postgres + pgvector + AGE | This is the moat per Blomfield. Memary and Ori-Mnemos are inspirations, not dependencies. The Brain has tenant-specific access patterns (RLS, per-skill memory scoping, episodic event sourcing) that no OSS package handles correctly out of the box. |
| Skill registry | Build (Go) | Anthropic's SKILL.md format is adopted as-is. The registry that stores, versions, signs, and serves Skills is ours. Rate-limited public catalog endpoints, private per-tenant fork, eval-on-deploy all live here. |
| Skill content (the Skills themselves) | Mirror + curate | We never write Skills from scratch when a high-quality public Skill exists. Mirror paperclipai/companies, harness-100, lenny-skills, gooseworks, CosmoBlk, Medical-Skills. Curate by department. Pay community contributors revenue share. |
| Paperclip | Fork the UX, build on our kernel | Paperclip is the single best public demonstration of how an AI company should look and feel — org chart, ticket system, budget caps, heartbeats, multi-company. The UX patterns are gold. The runtime underneath cannot be ours' (Paperclip is single-machine, ours is multi-tenant SaaS). So we adopt the design vocabulary, not the code. |
| n8n | Adopt as-is + write custom Galileo nodes | 47K stars, Fair-code license, 400+ integrations. We package n8n as the visual workflow surface. We contribute Galileo nodes (brain.read, agent.invoke, org.escalate) back upstream and ship them in our own n8n distribution. |
| Customer support inbox (Chatwoot) | Adopt + isolate (AGPL) | AGPL viral concerns are real but manageable: Chatwoot runs as its own service behind a documented API. We do not modify Chatwoot source. Frappe Helpdesk is the alternative for ERPNext customers. |
| CRM backend | Adopt Twenty CRM + NocoDB; fall back to building | Stage 2 picks the best fit per tenant. If neither maps cleanly to a tenant's process, we drop in NocoDB and let the Sales Skill define the schema. Galileo does not have an opinion about CRM data model. |
| Scheduling (Cal.com) | Adopt + isolate (AGPL) | Self-hosted Cal.com behind a thin gateway. Same AGPL containment as Chatwoot. |
| Vector DB (pgvector → Weaviate) | Adopt both, promote on demand | Default everyone to pgvector for simplicity. Promote a tenant to Weaviate only when they cross 10M vectors. Most never do. Weaviate is Go and self-hostable, so promotion is a config change, not a vendor swap. |
| Auth (Authgear / Supabase Auth) | Adopt | Supabase Auth ships with Stage 0. Authgear takes over for Stage 4 enterprise SSO/SAML/OIDC. Both are Go. We build no auth code of our own. |
| Secrets (Infisical) | Adopt | Per-tenant per-env secret rotation, audit log, no Vault operational complexity. MIT-licensed. |
| Observability (Opik for LLM, Prometheus for infra) | Adopt both | Opik is built specifically for LLM tracing, scores, regression. Prometheus + Grafana for everything else. Sentry self-hosted for app errors. |
| Billing (Stripe + custom meter) | Adopt Stripe + build the meter | Stripe is the payments rail. The metering layer that aggregates LiteLLM events into Stripe invoice items is ours, in Go, sitting next to LiteLLM. |
| Self-host installer (Coolify) | Adopt | Coolify is to Galileo what RDS is to Heroku — the operational layer customers expect. Our docs will assume Coolify for the SMB tier. |
| Web app (Next.js + Tailwind + shadcn/ui) | Adopt the stack, build the app | No reinvention. Standard Next.js 16 with App Router, shadcn/ui components, Tailwind. The differentiator is what the app does, not what it's built with. |
| Mobile (React Native + Expo) | Adopt + build | User explicitly named React Native as a preferred language. Expo Router gives us iOS + Android + web from one codebase. Hard requirement: the operator can approve agent actions from a phone in under five seconds. |
| Desktop (Tauri + Rust) | Adopt + build | When a Stage 2 power user asks for a desktop app that wraps the operator and has system tray notifications, Tauri ships in 2KB instead of Electron's 200MB. Rust shell, web UI inside. |
| WhatsApp gateway (open-bsp-api) | Adopt | Self-host WhatsApp Business Platform without Twilio's per-message markup. Important for non-US markets where WhatsApp is the operator surface. |
| Coding agents (OpenHands, Aider, Cline, Claude Code) | Adopt as customers, never build | Galileo does not compete with coding agents. Galileo is built using them. The 'no engineering department' constraint is not a limitation — it is a positioning advantage. |
| Unified data plane (Mirage) | Adopt at Layer 5 (agent-side) | Apache 2.0, OSS, v0.0.1 May 2026. Imported in-process by Python and TypeScript agents that want unified-filesystem abstraction across heterogeneous backends. Per-agent choice alongside discrete MCP servers. Mirage is **not** a kernel-side substrate (see `docs/closeouts/CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md` and `docs/decisions/0003-mirage-layer-relocation.md`). Stage 2 may re-evaluate placement if Mirage publishes a Go SDK or native server mode. |
| Onboarding Crew | Build (Python + Go) | Nothing in the public catalog runs the six-agent crew described in §3. The Connector / Crawler / Ingestion / Org-Mapper / Skill-Selector / QA pattern is Galileo's. This is the customer-facing front door, not a commodity. |
| Social media scraping (cporter202 list and similar) | Do not bundle. Opt-in Apify MCP only. | Affiliate-linked Apify actor lists are not OSS code to vendor in. Shipping LinkedIn / IG / TikTok scraping as a default feature creates ToS exposure (LinkedIn v. hiQ) and an unbounded maintenance burden. Right answer: one Apify MCP server, opt-in per tenant, explicit ToS acknowledgment, per-platform rate caps, audit logged. |
| Notification fan-out | Adopt (Apprise) | Apprise (Tom Doerr May 10 2026) covers 130+ notification services through one Python API. Cheaper than building per-channel notifier code. BSD-2 license. Replaces ad-hoc per-channel notifier code in the operator surface. |
| Destructive-action gating | Build (Go, Temporal signals) | No off-the-shelf component does this correctly. Galileo's destructive-action lockdown is a Temporal-signal-based approval router gated on a pre-write snapshot artifact recorded by the agent that performs the destructive operation. Mirage is one tool (Python agents may use `workspace.snapshot()`) for producing the artifact; agents using discrete connectors record per-source backups; Brain state has Postgres PITR. The router itself is kernel-side; the snapshot mechanism is agent-side. It is the architecture of trust — the thing that makes a B2B customer willing to give Galileo write scope at all. We build it, we own it, we test it on every release. |

# Appendix A. Curated Repo Index

The full inventory of OSS repositories referenced anywhere in this plan, organized by the layer where they enter the stack. Stars and dates reflect the April–May 2026 Tom Doerr archive plus the source document. This is the sourcing list for an engineer asked to 'go evaluate what's out there for X' — start here, then branch out.

## A.1 Layer 0 — Foundation

- **supabase/supabase** — github.com/supabase/supabase — Postgres + auth + storage + realtime + pgvector. The default backend.
- **authgear/authgear-server** — Go-native enterprise SSO/SAML/OIDC. Stage 4.
- **Infisical/infisical** — TS/Rust secrets management. MIT.
- **coollabsio/coolify** — Self-hosted Heroku/Vercel. The standard installer for SMB self-hosters.
- **kubernetes/helm + akuity/kargo** — Standard K8s GitOps for the managed cloud tier.
- **louislam/uptime-kuma** — Self-hosted status page and uptime monitor.

## A.2 Layer 1 — LLM Gateway

- **BerriAI/litellm** — One API for 100+ LLMs. The single chokepoint for cost and routing.
- **vllm-project/vllm** — Production-grade local inference. Continuous batching, paged attention, OpenAI-compatible.
- **ollama/ollama** — Dev/edge local inference. Same OpenAI shape as vLLM.
- **comet-ml/opik** — LLM tracing, scores, regression detection. Stage 0.
- **instructor-ai/instructor** — Structured outputs from LLMs via Pydantic schemas.
- **dottxt-ai/outlines** — Grammar-constrained generation. Guaranteed JSON / regex outputs.
- **stanfordnlp/dspy** — Programmatic prompt optimization. Stage 3.
- **confident-ai/deepteam** — LLM red-team / jailbreak testing. Stage 2.

## A.3 Layer 2 — Orchestration Kernel

- **temporalio/temporal** — Go-native durable execution engine. Stripe / Snap / Coinbase scale. Multi-region GA 2026.
- **nats-io/nats-server + nats-io/jetstream** — Sub-millisecond, multi-tenant message bus. Go.
- **langchain-ai/langgraph** — Stateful agent graphs. Production-grade.
- **crewAIInc/crewAI** — Role-based agent teams.
- **agno-agi/agno** — Fast simple multi-modal agents. 10x faster than LangChain for simple cases.
- **microsoft/autogen** — Multi-agent conversation framework. Reference implementation.
- **huggingface/smolagents** — Minimal code-first agent framework, ~1000 lines total.

## A.4 Layer 3 — Company Brain

- **pgvector/pgvector** — C extension for Postgres. Default vector store.
- **apache/age** — Graph database extension on Postgres. Default graph store at small/medium scale.
- **weaviate/weaviate** — Go-native vector DB at scale. Stage 3 promotion target.
- **DS4SD/docling** — IBM's high-fidelity PDF / Office to markdown converter.
- **microsoft/markitdown** — Universal file-to-markdown for LLMs.
- **unclecode/crawl4ai** — LLM-ready web scraping in Python.
- **mendableai/firecrawl** — Production-grade managed crawling. AGPL.
- **Vaibhavs10/insanely-fast-whisper** — Audio ingestion at 10–20x Whisper speed.
- **umarbutler/semchunk** — Semantic chunking for RAG.
- **HKUDS/RAG-Anything** — Multimodal RAG with tables, figures, charts.
- **kingjulio8238/Memary** — Long-term memory pattern for autonomous agents.
- **whyhow-ai/knowledge-graph-studio** — Graph admin UI for the Brain.
- **adbar/trafilatura** — Single-page web extractor.
- **Unstructured-IO/unstructured** — Document parser ecosystem.
- **HKUDS/RAG-Anything** — Multimodal RAG; production-grade.
- **strukto-ai/mirage** — Unified VFS for AI agents. Apache 2.0. Mounts S3, GDrive, Slack, Gmail, GitHub, Notion, Linear, Postgres, MongoDB, SSH side-by-side. Snapshot/clone/rollback workspaces. Python ≥ 3.12 + TypeScript SDKs. Launched May 6 2026. THE substrate for the Onboarding Crew's crawl, pending Stage 0 probe.

## A.5 Layer 4 — Skills

- **anthropics/claude-code-skills** — Official SKILL.md format spec.
- **paperclipai/companies** — 16 pre-built AI companies, 440+ specialized agents (Apr 2026 Tom Doerr feed).
- **revfactory/harness-100** — 100 agent teams across 10 domains (May 4 2026).
- **garrytan/gstack** — 15 specialist Skills built by Garry Tan / YC. Source-document anchor.
- **RefoundAI/lenny-skills** — 86 product-management Skills (Apr 29 2026).
- **gooseworks/ai-goose-skills** — Growth and GTM Skills (Apr 27 2026).
- **CosmoBlk/email-marketing-bible** — Comprehensive email marketing skill set (Apr 29 2026).
- **FreedomIntelligence/OpenClaw-Medical-Skills** — 869 medical AI Skills (May 3 2026). Available as an optional vertical pack in Stage 4.
- **Karanjot786/agent-skills-cli** — 175,000+ agent Skills indexed. Discovery surface.
- **Microck/ordinary-claude-skills** — Categorized Claude Skill library with search.
- **msitarzewski/agency-agents** — Source-document repo: 8-department agency template.
- **karpathy/autoresearch** — The autoresearch loop pattern. Source-document anchor.
- **paperclipai/paperclip** — Source-document anchor. UX vocabulary for Galileo's operator surface.
- **promptfoo/promptfoo** — Skill eval suite.
- **anthropics/skill-check** — SKILL.md linter.

## A.6 Layer 5 — Integrations

- **modelcontextprotocol/specification** — MCP standard. Wire format for every Galileo tool.
- **punkpeye/awesome-mcp-servers** — 500+ ready-made MCP servers.
- **n8n-io/n8n** — Visual workflow automation, 400+ apps. Self-hosted.
- **ComposioHQ/composio** — Managed connector catalog.
- **microsoft/playwright-mcp** — Real browser as a tool for agents.
- **browser-use/browser-use** — Supervised browser automation.
- **calcom/cal.com** — Self-hosted scheduling. AGPL.
- **googleworkspace/cli** — Google Workspace CLI with AI-agent skills. Source-document anchor.
- **apify/actor-platform** — Managed scraping infrastructure with thousands of actors. Used via opt-in MCP server, never bundled by default. ToS-explicit per-tenant.
- **caronc/apprise** — One API for 130+ notification services. BSD-2. Tom Doerr May 10 2026.

## A.7 Layer 6 — Department Modules

- **brightbean-studio/brightbean** — OSS Buffer/Hootsuite for social management. Marketing department.
- **jay-sahnan/signal** — OSS Apollo/Outreach for sales intel. Sales department.
- **chatwoot/chatwoot** — OSS Intercom/Zendesk. Support department. AGPL.
- **frappe/helpdesk** — Frappe/ERPNext-aligned helpdesk. Support department. AGPL.
- **arc53/DocsGPT** — OSS knowledge base + chat. Support department.
- **twentyhq/twenty** — Modern OSS CRM. Sales department. AGPL.
- **nocodb/nocodb** — OSS Airtable. Flexible CRM/ops backend.
- **invoiceninja/invoiceninja** — OSS invoicing. Finance department.
- **jameskokoska/Cashew** — OSS expense tracking. Finance department.
- **mendableai/fire-enrich** — Email-list to company-data enrichment. Sales/Marketing.
- **frappe/erpnext** — Full ERPNext for HR / payroll / inventory. Finance + HR.
- **google-labs-code/design.md** — DESIGN.md spec for design-as-Skill. Design department.
- **Hawksight-AI/semantica** — Context graphs and decision intelligence. Source-document anchor for the Brain.

## A.8 Layer 7 — Operator

- **vercel/next.js** — Web admin app framework.
- **shadcn-ui/ui** — Component system. Tailwind-based.
- **expo/expo** — React Native build pipeline. iOS + Android + web from one codebase.
- **tauri-apps/tauri** — Rust desktop shell. 2KB vs Electron's 200MB.
- **slackapi/bolt-js** — Slack bot framework.
- **dotta/clawputer / opencomputer.dev** — Source-document anchor for the Telegram-as-operator pattern.
- **open-bsp/api** — Self-hosted WhatsApp Business gateway.

## A.9 Cross-cutting

- **prometheus/prometheus + grafana/grafana** — Infra metrics.
- **PostHog/posthog** — Product analytics.
- **plausible/analytics** — Privacy-respecting web analytics.
- **getsentry/self-hosted** — Error tracking.
- **keephq/keep** — AIOps + alert management.
- **All-Hands-AI/OpenHands, Aider-AI/aider, cline/cline, anthropics/claude-code** — The coding agents Galileo's team uses to build Galileo. Not products Galileo competes with.

# Appendix B. Stage 0 Starter Stack

A literal docker-compose.yml that an engineer can copy on day one to bring up the entire Stage 0 stack on a single VM. This is not a production deployment; it is the smallest set of services that lets a 'Hello, Agent' demo run end-to-end with durability, tracing, and budget enforcement. Stages 1+ replace the single-VM Coolify deploy with a Helm chart, but the service inventory is the same.

```
# galileo-stage0/docker-compose.yml
# Single-VM Stage 0 stack. Run on a Hetzner CCX23 / DO 4-vCPU box.
version: '3.9'
 
services:
  postgres:
    image: supabase/postgres:15.6.1.135
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes: [pgdata:/var/lib/postgresql/data]
    ports: ['5432:5432']
 
  supabase:
    image: supabase/studio:latest
    depends_on: [postgres]
    ports: ['3000:3000']
 
  temporal:
    image: temporalio/auto-setup:1.27
    environment:
      DB: postgresql
      DB_PORT: 5432
      POSTGRES_USER: postgres
      POSTGRES_PWD: ${POSTGRES_PASSWORD}
      POSTGRES_SEEDS: postgres
    depends_on: [postgres]
    ports: ['7233:7233']
 
  temporal-ui:
    image: temporalio/ui:2.32
    environment: { TEMPORAL_ADDRESS: temporal:7233 }
    depends_on: [temporal]
    ports: ['8080:8080']
 
  nats:
    image: nats:2.10-alpine
    command: '-js -sd /data'
    volumes: [natsdata:/data]
    ports: ['4222:4222', '8222:8222']
 
  litellm:
    image: ghcr.io/berriai/litellm:main-stable
    environment:
      DATABASE_URL: postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/litellm
      LITELLM_MASTER_KEY: ${LITELLM_MASTER_KEY}
      OPENAI_API_KEY: ${OPENAI_API_KEY}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY}
    depends_on: [postgres]
    ports: ['4000:4000']
 
  opik:
    image: comet/opik:latest
    depends_on: [postgres]
    ports: ['5173:5173']
 
  ollama:
    image: ollama/ollama:latest
    volumes: [ollama:/root/.ollama]
    ports: ['11434:11434']
 
  galileo-gateway:        # custom Go service
    image: galileo/gateway:0.1.0
    environment:
      LITELLM_URL: http://litellm:4000
      LITELLM_KEY: ${LITELLM_MASTER_KEY}
      DATABASE_URL: postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/galileo
    depends_on: [litellm, postgres]
    ports: ['8001:8001']
 
  galileo-agent-runner:   # custom Go service
    image: galileo/agent-runner:0.1.0
    environment:
      TEMPORAL_HOST: temporal:7233
      NATS_URL: nats://nats:4222
      GATEWAY_URL: http://galileo-gateway:8001
      DATABASE_URL: postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/galileo
    depends_on: [temporal, nats, galileo-gateway]
 
  galileo-web:            # Next.js admin
    image: galileo/web:0.1.0
    environment:
      NEXT_PUBLIC_GATEWAY_URL: http://galileo-gateway:8001
      NEXT_PUBLIC_TEMPORAL_UI: http://localhost:8080
    depends_on: [galileo-gateway]
    ports: ['3001:3001']
 
volumes:
  pgdata: {}
  natsdata: {}
  ollama: {}
```

With the above committed to the galileo-stage0 repo and a one-line .env file, a fresh Ubuntu 24.04 VM goes from 'docker compose up -d' to a working agent in under fifteen minutes. The Stage 0 success gate is exactly this — and a 30-minute walkthrough video that any senior engineer can complete without help.

> **End of plan**
>
> This document is Galileo OS v0.1 of the infrastructure plan. It is intentionally opinionated. Every decision in here is reversible — but the cost of reversing a decision goes up sharply once code ships, so the goal of this document is to make the irreversible-feeling decisions explicit before any code is written.
>
> The single most important strategic input — Tom Blomfield's Company Brain essay — should be re-read before every stage gate. The Brain is the moat. Everything else is plumbing.
