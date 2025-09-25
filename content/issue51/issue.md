---
title: "AllowedTools vs YOLO mode: Secure But Powerful Agentic Engineering"
publishDate: "2025-09-25T09:00:00-07:00"
pubDatetime: 2025-09-25T09:00:00-07:00
newsletterSubject: "AllowedTools vs YOLO mode: Secure But Powerful Agentic Engineering"
tags: ["Claude Code", "Security", "YOLO Mode", "Allowed Tools", "Agentic Engineering"]
socialMediaHashtags: "#ClaudeCode #Security #AgenticEngineering #DeveloperTools"
contentID: "issue51"
description: "Recently, I've defaulted to using my coding agents in YOLO mode. I found a better way to balance security and ease of use."
---

![AllowedTools vs YOLO mode: Secure But Powerful Agentic Engineering](./banner.png)

Recently, I've defaulted to using my coding agents in YOLO mode. I found a better way to balance security and ease of use.

Once you get the hang of agentic coding, it can feel like babysitting. Can I read this file? Can I search these directories? Everything has to be allowed individually by default. The easiest fix is to switch to YOLO mode. Instead of starting claude in the terminal, start claude --dangerously-skip-permissions. This allows your agent to do everything: read all the files, delete all the files, commit to every repository on your hard disk. Even connecting to production servers and databases using your SSH keys. YOLO mode is the right name, real accidents happened.

But YOLO mode has limitations too. I started to install Claude on my managed servers. It's helpful for boring server administration tasks. Unfortunately, Claude doesn't work in YOLO mode when you're the root user, which is typical for cloud machines. I'm not sure if I agree with Anthropic's limitation, since this can be less dangerous than running Claude on my private machine with all my private data in YOLO mode.

Fortunately, better options are emerging. One I like is [allowed tools](https://docs.claude.com/en/docs/claude-code/sdk/custom-tools#configuring-allowed-tools). This gives the agent fine-grained controls on what he can do on his own and what not. Together with the [slash commands](https://ontree.co/en/blog/2025/custom-slash-commands-a-field-trip-with-claude-code/), I wrote about last week, this is a powerful combination. Similar to the [dotfiles](https://wiki.archlinux.org/title/Dotfiles) that many developers use for a familiar environment on new machines, I can imagine checking out a claude-tools repository with custom slash commands for repeating tasks. And also including allowedTools for uninterrupted execution.

Disclaimer: I haven't built this yet. Hopefully, I'll have a demo for you in the next weeks!

<!--LINKS_SEPARATOR-->

### The death of the corporate job is an evergreen topic

- **URL:** https://thestillwandering.substack.com/p/the-death-of-the-corporate-job
- **MyTake:** The death of the corporate job is an evergreen topic, of course. But so far, people lacked to see a realistic alternative to their corporate pointlessness. It really has become easier to run a one-man-show with AI. Let's see where this goes!
- **Keyword:** link

---

### Absolutely nailed it, I agree!

- **URL:** https://openshovelshack.com/blog/let-it-marinate
- **MyTake:** "The main bottleneck for building software (and other systems) is the time required to let the ideas bounce around in our head until we feel that we have a good grasp of what direction to take." Absolutely nailed it, I agree!
- **Keyword:** link

---

### Hardware is painful compared to software

- **URL:** https://www.youtube.com/watch?v=N5xhOqlvRh4
- **MyTake:** I sometimes watch these hardware videos, just to remind myself how painful hardware is compared to software. The main message is a bit hidden: It rarely makes sense to cluster computers for a task. Each node must be able to process requests individually, leading to parallelization. This was true before and is true with LLM inference. If you need to run a bigger LLM model than your system can hold, you're better off buying a bigger system that can hold it.
- **Keyword:** link

<!--PRINT_SEPARATOR-->

## What to Print This Week

### Garden Races - Create Your Own RC World

![Garden Races - Create Your Own RC World](https://makerworld.bblmw.com/makerworld/crowdfunding/20250911/3926511493/10d7c412ad0623a7.jpg?x-oss-process=image%2Fresize%2Cw_1920%2Fformat%2Cwebp)

Now this makes a lot of sense. Instead of long lists with items to buy off Aliexpress, you order one CyberBrick [electronics kit](https://eu.store.bambulab.com/products/cyberbrick-hardware-kit) for 46 Euros (82 Euros for a double kit) and 18 Euros for this set of files to print. Lego MindStorm on steroids, if you can still remember. And much more sustainable, as you can reuse the electronics kit with other models later on.

[visit model page](https://makerworld.com/en/crowdfunding/76-garden-races-create-your-own-rc-world)

---

### CyberBrick Official Truck

![CyberBrick Official Truck](https://makerworld.bblmw.com/makerworld/model/US23a7c410ecc57d/design/2025-05-09_db9e1374a5e018.jpg)

The original CyberBrick Truck, this one is free.

[visit model page](https://makerworld.com/en/models/1396031-cyberbrick-official-truck?from=search#profileId-1447043)

---

### Modular CyberBrick Race Track Builder

![Modular CyberBrick Race Track Builder](https://makerworld.bblmw.com/makerworld/model/US9138b24569e736/design/2025-08-18_967d8547942778.jpg)

And a Race Track, though this car isn't a racer at all.

[visit model page](https://makerworld.com/en/models/1565356-modular-cyberbrick-race-track-builder?from=search#profileId-1813782)

<!--FOOTER_SEPARATOR-->

## Hi 👋, I'm Stefan!

This is my weekly newsletter about new technology hypes in general and AI in specific. Feel free to forward this mail to people who should read it. If this mail was forwarded to you, please [subscribe here](https://liquid.engineer).

[https://liquid.engineer](https://liquid.engineer)

Stefan Munz, www.stefanmunz.com
