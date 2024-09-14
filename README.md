[![License](https://img.shields.io/badge/license-MIT-blue)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active)

# clip4llm 🚀

Why waste time fiddling with files when you could be crafting prompts for your LLM? **clip4llm** takes the hassle out of manually copying file content into ChatGPT or any other LLM. This isn’t just a clipboard helper; it’s a speed boost for your brain. You point, you copy, and you’re instantly back to the fun part—prompting.

No more juggling files or struggling with hidden formats. Whether it’s code snippets, config files, or that critical .env file, **clip4llm** does the heavy lifting so you can focus on what really matters: getting that sweet AI-generated insight. Grab your content and get back to prompting—faster, easier, and smarter.

## 🌟 Features

- **Wildcard File Handling:** Why limit yourself to one file when you can have them all? Use wildcards to gather your files (e.g., `*.md`, `*.yaml`) and let **clip4llm** do the rest.
- **Hidden Gems:** Hidden files aren’t left behind. Grab those `.env` secrets like a pro.
- **Size Matters:** File too big? Not a problem. Set a size limit and skip the heavyweights. Default: 32KB, because nobody needs a novel-length paste job consuming your precious context window.
- **Mind the Megabyte:** Output over 1MB? Boom! That is too big so nope, not happening.
- **Binary Exclusion:** ChatGPT doesn’t speak binary—leave those files out automatically.
- **Config Magic:** Drop a `.clip4llm` config in your home directory or your project folder and forget about the command-line—your preferences are locked and loaded.
- **Verbose Mode:** Want to see what’s going on behind the curtain? Crank up the verbosity and feel like a hacker.

## 🔧 Installation

### Option 1: Build from Source

You’ve got Go installed, right? If not, ask ChatGPT how to instally it and let's get going...

1. Clone the magic:

   ```bash
   git clone https://github.com/unitvectory-labs/clip4llm.git
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

### Option 2: Download the Binary

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

- `--delimiter` – Customize how each file is wrapped. Default is tripple `'s because Markdown rocks, but make it whatever you like:

  ```bash
  clip4llm --delimiter="<<<END>>>"
  ```

- `--max-size` – Keep things trim. Files over this size (in KB) are skipped:

  ```bash
  clip4llm --max-size=8
  ```

- `--include` – Specific about what you want? Use a comma-separated list of filenames or patterns, those files and folders starting with a . are automatically skipped so this is your opportunity to bring some of them back into the fold:

  ```bash
  clip4llm --include=".github,*.env"
  ```

- `--exclude` – Want some files to disappear? Exclude 'em with style:

  ```bash
  clip4llm --exclude="LICENSE,*.md"
  ```

- `--verbose` – Feeling nosy? Get the full play-by-play of what’s happening:

  ```bash
  clip4llm --verbose
  ```

### 🔥 Pro Tip Combos

- **Include Hidden Directory**: Just the `.github` folder, please:

  ```bash
  clip4llm --include=".github"
  ```

- **Exclude Markdown Files**: No `.md` files here:

  ```bash
  clip4llm --exclude="*.md"
  ```

- **Mix & Match**: Include `.env`, ditch `.md`:

  ```bash
  clip4llm --include="*.env" --exclude="*.md"
  ```

- **Customize Your Flow**: Use a custom delimiter and max file size, keep that LLM happy:

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
```
