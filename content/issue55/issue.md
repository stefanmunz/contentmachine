---
title: "Custom Slash Commands, Part 3: The Installer"
publishDate: "2025-10-16T09:00:00-07:00"
pubDatetime: 2025-10-16T09:00:00-07:00
newsletterSubject: "Custom Slash Commands, Part 3: The Installer"
tags:
  [
    "Claude Code",
    "Automation",
    "Custom Commands",
    "Developer Tools",
    "Installers",
  ]
socialMediaHashtags: "#ClaudeCode #Automation #DeveloperTools #Installers"
contentID: "issue55"
description: "Three patterns that turn installation scripts into self-fixing, robust automation. This is how I build installers and other scripts that improve themselves with every run."
---

![Custom Slash Commands, Part 3: The Installer](./banner.png)

In [part one](https://www.stefanmunz.com/posts/2025/custom-slash-commands-a-field-trip-with-claude-code/custom-slash-commands-a-field-trip-with-claude-code), custom slash commands automated repetitive tasks but gave inconsistent results. In [part two](https://www.stefanmunz.com/posts/2025/custom-slash-commands-part-2-from-convenience-to-100-repeatability/custom-slash-commands-part-2-from-convenience-to-100-repeatability), we found the fix: separate the conversational prompt from a deterministic script.

Now, the finale. After weeks of hardening this approach, I've distilled it into three patterns that transform scripts into powerful self-fixing installers. Here's what I learned building the TreeOS production setup.

## The Three Patterns

### 1. The Self-Fixing Loop

The old way: run the script, watch it fail, open the file, find line 52, guess a fix, save, run again. High-friction context switching.

The new way: I run the script, it fails on an edge case, and I say: "Claude, that error on line 52—the script needs to back up existing directories instead of failing. Fix it."

Claude has the full context: the command, the code, and the failed output. It updates the script immediately. The script hardens with each real-world failure. This tight feedback loop is the fastest way to build robust automation. TreeOS will be open source. Users can run the install script and contribute a pull request if they encounter an edge case. Can't wait to see this in real life.

### 2. Soft Front Door + Hard Engine

Every installer consists of two side-by-side files:

- Soft Front Door (`.md`): treeos-setup-production.md

- Hard Engine (`.sh`): treeos-setup-production-noconfirm.sh

The markdown contains the Claude Code prompt, the conversational layer. It explains what's about to happen, checks prerequisites, and asks for confirmation. It's flexible and human-friendly.

The shell script is the deterministic engine. It takes inputs and executes precise commands. No ambiguity, no improvisation, 100% repeatable.

This separation is crucial. Claude can safely modify the conversation in the front door without breaking the logic in the engine. The naming convention makes the relationship obvious.

### 3. The Graceful Handoff

My scripts run on machines where I have sudo and on servers where Claude Code doesn't. Both need to work seamlessly.

The pattern: check if sudo is available without a password prompt.

```bash

sudo -n true 2>/dev/null && echo "SUDO_AVAILABLE" || echo "SUDO_REQUIRED"

```

If sudo requires a password, the front door hands off cleanly:

```

⚠️ This script requires sudo privileges.

Claude Code cannot provide passwords for security reasons.

I've prepared everything. Run this one command:

cd ~/repositories/ontree/treeos

sudo ./.claude/commands/treeos-setup-production-noconfirm.sh

Paste the output back here, and I'll verify success.

```

Claude does 95% of the work, then asks me to handle the one step it can't. Perfect collaboration.

## The Real-World Result

These three patterns produced my TreeOS production installer. It's now 600+ lines and handles:

- OS detection (Linux/macOS) and architecture

- Downloading the correct binary from GitHub releases

- Creating system users with proper permissions

- Optional AMD ROCm installation if a GPU is detected

- Service setup (systemd/launchd) and verification

When something breaks on a new platform, the self-fixing loop makes improvements trivial. I've hardened this across dozens of edge cases without dreading the work.

## Why This Changes Everything

Traditional README files demand a lot from the user. They push the cognitive load onto users: identify your platform, map generic instructions to your setup, debug when it breaks.

This flips the script. Instead of static documentation describing a process, we have executable automation that performs it.

But this isn't just about installers. Apply these patterns to any complex developer task:

- /setup-dev-environment clones repos, installs tools, and seeds databases

- /run-migration backs up production, runs the migration, and rolls back on failure

- /deploy-staging builds containers, pushes to registries, and updates Kubernetes

We're moving from documentation that describes to automation that executes, with AI as the safety net and co-pilot. This is the future of developer experience: reducing friction by automating complex workflows around code.

With the explosion of AI tools, setup complexity is a real barrier. These patterns are one step towards changing that.

<!--LINKS_SEPARATOR-->

### Link Title Here

- **URL:** https://example.com
- **MyTake:** Your take on this link
- **Keyword:** link

<!--PRINT_SEPARATOR-->

## What to Print This Week

### Print Model Title

![Print Model Title](./image.jpg)

Description of what makes this print interesting.

[visit model page](https://makerworld.com/en/models/example)

<!--FOOTER_SEPARATOR-->

## Hi 👋, I'm Stefan!

This is my weekly newsletter about technology becoming more fluid and adaptive - from rigid software to liquid tools that shape themselves to our needs. Feel free to forward this mail to people who should read it. If this mail was forwarded to you, please subscribe here, it's always 1 mail per week. https://liquid.engineer.

Stefan Munz, www.stefanmunz.com
