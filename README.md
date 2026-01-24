[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/clip4llm.svg)](https://github.com/UnitVectorY-Labs/clip4llm/releases/latest) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active) [![Go Report Card](https://goreportcard.com/badge/github.com/UnitVectorY-Labs/clip4llm)](https://goreportcard.com/report/github.com/UnitVectorY-Labs/clip4llm)

# clip4llm 🚀

Why waste time fiddling with files when you could be crafting prompts for your LLM? **clip4llm** takes the hassle out of manually copying file content into ChatGPT or any other LLM. This isn’t just a clipboard helper; it’s a speed boost for your brain. With one simple command it grabs that text to feed those hungry LLMs the tokens they need and you are back to the fun part... prompting.

No more juggling files or flipping back and forth between text editors. Whether it’s code snippets, config files, or that critical .env file, **clip4llm** does the heavy lifting of copying all of those files at once so you can focus on what really matters: getting that sweet AI-generated insight.

Does **clip4llm** actually interact with ChatGPT or any other LLM API directly, absolutely not! That is your job. All it does is copy those text files you are working straight to your clipboard so you can do the real work of pasting them into your LLM of choice. Spend all the time this saves you sending more prompts to that LLM getting back those hallucinatory insights you crave.

## 🌟 Features

- **Hidden Gems:** Hidden files aren’t included by default, but you can include them to grab those `.env` secrets like a pro.
- **Size Matters:** File too big? Not a problem. Set a size limit and skip the heavyweights. Default: 32KB, because nobody needs a novel-length paste job consuming your precious context window.
- **Mind the Megabyte:** Output over 1MB gathered? Boom! That is too big so nope, not happening.
- **Binary Exclusion:** ChatGPT doesn’t speak binary—leave those files out automatically.
- **Config Magic:** Drop a `.clip4llm` config in your home directory or your project folder and forget about the command-line—your preferences are locked and loaded.
- **Scoped Configs:** Need different rules for different directories? Drop `.clip4llm` files in subdirectories for fine-grained control—like `.gitignore` but for your LLM context.
- **Verbose Mode:** Want to see what’s going on behind the curtain? Crank up the verbosity and feel like a hacker.

## 🔧 Installation

### Option 1: Install using Go

You’ve got Go installed, right? If not, ask ChatGPT how to install it and let's get going...


1. Install it the easy way:

   ```bash
   go install github.com/UnitVectorY-Labs/clip4llm@latest
   ```

2. Make sure those Go bins are in your path:

   ```bash
   export PATH=${PATH}:$(go env GOPATH)/bin
   ```

3. Check your setup (just because):

   ```bash
   clip4llm --help
   ```

### Option 2: Build from Source

You’ve got Go installed, but want to manually build everything? Good for you, let's get started...

1. Clone the magic:

   ```bash
   git clone https://github.com/UnitVectorY-Labs/clip4llm.git
   cd clip4llm
   ```

2. Build it like a boss:

   ```bash
   go build -o clip4llm
   ```

3. Make it global:

   ```bash
   mv clip4llm /usr/local/bin/
   ```

4. Check your setup (just because):

   ```bash
   clip4llm --help
   ```

### Option 3: Download the Binary

1. Cruise over to [clip4llm Releases](https://github.com/UnitVectorY-Labs/clip4llm/releases) and snag the latest version for your OS—Mac, Linux, Windows, whatever team you roll with.
2. Unzip the file (or untar it if it's a tar.gz, whatever, you know what to do).
3. Move that binary into your PATH or let it chill in your downloads folder forever. Your call.
4. Fire up **clip4llm** and you're golden.

## 💻 Usage

Time to flex. Navigate into your project directory and run:

```bash
clip4llm
```

This instantly grabs all non-hidden, non-binary files (under 32KB) in your current directory and copies them straight to your clipboard, ready for pasting into ChatGPT like a legend.

### Command-Line Options That Matter

- `--delimiter` – Customize how each file is wrapped. Default is triple `'s because Markdown rocks, but make it whatever you like:

  ```bash
  clip4llm --delimiter="<<<END>>>"
  ```

- `--max-size` – Need fatter files, up that max-size (KB) to something bigger if you have context window to burn:

  ```bash
  clip4llm --max-size=8
  ```

- `--include` – By default those .files and .folders are left out, if you want them you need to specify them here:

  ```bash
  clip4llm --include=".github,*.env"
  ```

- `--exclude` – Some files wasting those tokens, exclude 'em with style:

  ```bash
  clip4llm --exclude="LICENSE,*.md"
  ```

- `--verbose` – Feeling nosy? Get the full play-by-play of what’s happening:

  ```bash
  clip4llm --verbose
  ```


- `--no-recursive` – Only want files right here, not buried down three directories deep? This flag keeps it simple and stays in the current directory:

  ```bash
  clip4llm --no-recursive
  ```

### 🔥 Pro Tip Combos

- **Include Hidden Directory**: Maybe you need to debug that GitHub Action, include those files easily:

  ```bash
  clip4llm --include=".github"
  ```

- **Exclude Markdown Files**: Your empty markdown files not helping the LLM out, you can leave those out:

  ```bash
  clip4llm --exclude="*.md"
  ```

- **Exclude `node_modules`**: Because your 20 line TypeScript project pulled in 500MB of libraries and no way that LLM needs all that:

  ```bash
  clip4llm --exclude="node_modules"
  ```

- **Customize Your Flow**: Your file's may have some of those triple `'s so if you want different separators, you can set them:

  ```bash
  clip4llm --delimiter="<<<FILE>>>" --max-size=64
  ```

## ⚙️ Configuration Like a Boss

Set it once, and forget it. Place a `.clip4llm` file in your home directory (`~/.clip4llm`) or project directory (`pwd/.clip4llm`), and **clip4llm** will respect your preferences.

Sample `.clip4llm` file:

```properties
delimiter=```
max-size=32
include=.github,*.env
exclude=LICENSE,*.md
no-recursive=false
```

## 📂 Scoped Configuration (Directory-Level Rules)

Need different rules in different corners of your repo? Drop a `.clip4llm` anywhere and it becomes the law for that folder and everything below it. Like `.gitignore`, but for feeding your LLM and keeping your context window from exploding.

### How it works (no fluff)

- `~/.clip4llm` loads first (global vibes)
- the `.clip4llm` in the folder you run `clip4llm` from loads next (project vibes)
- then every time `clip4llm` walks into a subfolder, if it finds a `.clip4llm`, it stacks it on top (local vibes)
- CLI flags still dunk on all of it, because obviously

Closest folder wins for single values:
- `delimiter`
- `max-size`
- `no-recursive`

Patterns pile up for the list stuff:
- `include` adds more “yes”
- `exclude` adds more “no”

### Example

```
project/
├── .clip4llm              # exclude=*.md
├── api/
│   └── .clip4llm          # max-size=128
└── frontend/
    └── .clip4llm          # exclude=*.css
```

What happens:
- `*.md` gets yeeted everywhere
- `api/` can slurp bigger files (128KB) without crying
- `frontend/` also ignores CSS, because nobody wants 80,000 lines of “just vibes” styling

### Patterns

Your `include` / `exclude` can match either:
- file or folder names (`*.md`, `node_modules`)
- paths from the project root (`docs/*`, `api/specs/*.json`)

### Verbose mode

Run with `--verbose` and it’ll tell you when it finds and uses scoped configs so you can feel powerful and in control (even if you are not).