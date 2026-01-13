---
title: "The Case for TreeOS, Part 1"
publishDate: "2026-01-02T09:00:00-07:00"
pubDatetime: 2026-01-02T09:00:00-07:00
newsletterSubject: "The Case for TreeOS, Part 1"
tags: ["Newsletter", "TreeOS", "AI", "LLM", "OnTree", "Local AI"]
socialMediaHashtags: "#Newsletter #TreeOS #AI #LLM #OnTree #LocalAI"
contentID: "issue63"
description: "Why the gap between local and cloud models matters, and why TreeOS bets on the mid-sized tier."
---

![The Case for TreeOS, Part 1](./banner.png)

In December, I released the first beta version of my small operating system, [TreeOS](https://github.com/ontree-co/treeos). It is freely available on GitHub and is accompanied by a B2B solution called [OnTree.co](https://ontree.co/) that lets you run internal or private AI-powered apps and skills 100% locally.

I am not happy with the software quality yet, but looking back, this is not surprising. When I started working with agents, it made sense to start with the most extreme approach: writing the entire code base with agents. Unsurprisingly, agentic engineering is very different from classical software engineering. I have learned a lot, as you can see in the result. I will summarize my learnings in a [conference talk](https://www.case-conf.com/session/the-claude-code-wars-and-the-return-of-the-software-spec) next week in Berlin, and probably in one of our upcoming Agentic Coding Meetups in Hamburg as well. I also have a much clearer picture of how to improve the code base.

This is the first in a series about the reasons that led me to build TreeOS.

## The LLM Gap

There is a gap between what current computers and smartphones can do locally and what is possible in the cloud with massive GPUs. Here is a rough classification:

**Mobile models (2B-8B)**

There are some excellent models in this size class, often highly specialized for one use case. For example, I use these models for speech-to-text or text-to-speech locally. Modern smartphones have up to 16 GB of memory, so the operating system and these models can run in parallel. From a battery perspective, it is important that these models only run for short periods of time, or your battery drains too fast.

**Local models (8B-32B)**

This is the classic local model class. There are a huge number of optimized models. Users of these models typically have a powerful gaming graphics card with 16-32 GB of RAM. This allows them to use their CPU normally and run long-running tasks on the graphics card.

**Cloud models (200B and up)**

The exact sizes of the major cloud models are not disclosed. One of the leading open-source models is [Kimi K2 Thinking](https://unsloth.ai/docs/models/kimi-k2-thinking-how-to-run-locally), which has 1 trillion parameters. This equates to a requirement of roughly 250 GB of RAM. Besides a maxed-out Mac Studio with 500 GB of RAM, I am not aware of any off-the-shelf hardware solution that can run these models locally.

**The gap (32B-200B)**

This gap is not receiving much attention at the moment, but I am optimistic that this will change in 2026. We have already seen that mixture-of-expert models, such as GPT-OSS 120B, are highly capable and can run at reasonable speeds. Also, computers in this class can run a 32B model and hold a lot of context. This is important for agentic workflows, where the agent must build a mental map of the problem, whether it is code-related or not. However, as this memory also needs to be stored in graphics RAM, consumer graphics cards are not suitable for these workloads.

With the AMD AI 300 and 400 series, affordable machines are now available that are ideally suited to long-running asynchronous tasks.

The first bet of TreeOS is that this niche will grow in 2026.

<!--LINKS_SEPARATOR-->

### What I learned this week

- **URL:** https://www.youtube.com/watch?v=JdWgAdBHcHc
- **MyTake:** I made my very first YouTube video showcasing TreeOS setup on a VPS. It is an easy way to test-run the system, and making the video was more fun than expected. I went with German audio and English subtitles. What do you think?
- **Keyword:** video

---

- **URL:** https://www.youtube.com/watch?v=wAIzlGwEAO0
- **MyTake:** It seems that vLLM started supporting Strix Halo. This is good news, as Ollama is no longer state of the art. Good times for local inference.
- **Keyword:** video

---

- **URL:** https://www.tomshardware.com/pc-components/gpus/nvidia-launches-vera-rubin-nvl72-ai-supercomputer-at-ces-promises-up-to-5x-greater-inference-performance-and-10x-lower-cost-per-token-than-blackwell-coming-2h-2026
- **MyTake:** If you are wondering why RAM prices are rising so much, NVIDIA's new Vera line supports up to 1.5 TB of RAM. Data center owners are willing to pay more, at least as long as the bubble holds.
- **Keyword:** link

---

- **URL:** https://www.youtube.com/watch?v=IAnFIan6Svo
- **MyTake:** It looks like the reign of dual boilers for home espresso machines is ending. Modern machines use thick-film heaters. Funny how the internals get simpler, but the price tag of those 20 kg classics stays.
- **Keyword:** video

---

- **URL:** https://www.youtube.com/shorts/0EA5XNEA7_4
- **MyTake:** 3D printing a benchy takes 20 minutes on my printer. Here it is done in under 2 minutes with dry-ice cooling. Hilarious prototype, I hope they keep going.
- **Keyword:** video

<!--PRINT_SEPARATOR-->

## What to Print This Week

### Zoetropic Cipher

A spinning top that hides a secret message. You can customize the message and optionally add an LED because why not.

![Zoetropic Cipher](https://makerworld.bblmw.com/makerworld/model/US8e402ecc71b5db/design/2025-11-10_5fa461e3530d.png)

[visit model page](https://makerworld.com/de/models/1849119-top-secret-puzzle-of-light-science-toy?from=recommend#profileId-1984255)

<!--FOOTER_SEPARATOR-->

## Hi 👋, I'm Stefan!

This is my weekly newsletter about technology becoming more fluid and adaptive - from rigid software to liquid tools that shape themselves to our needs. Feel free to forward this mail to people who should read it. If this mail was forwarded to you, please subscribe here, it's always 1 mail per week. https://liquid.engineer.

Stefan Munz, www.stefanmunz.com
