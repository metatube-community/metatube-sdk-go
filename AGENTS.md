# AGENTS.md

## Providers

### Get data the way the site's frontend does, never by guessing patterns

Every value a provider returns (metadata, cover, thumbnail, preview images, sample video URLs, and so on) must come from what the site's frontend actually uses: the page HTML, its embedded data (ld+json, RSC payload, inline scripts), or the APIs that the frontend calls.

- **Never derive data from observed patterns**, e.g., building an image or video URL with a format string, or rewriting a thumbnail URL into a full-size one. A guessed URL that returns 200 today is not evidence; the site may stop serving it at any time.
- **A derivation is allowed only when the site's frontend code does the same derivation.** In that case:
  - Put it in a dedicated function in the provider, not inline in a scraping callback.
  - Port the frontend logic faithfully, branch by branch, without simplifying or extending it.
  - Above the function, add a `Ref:` comment with the URL of the JS file (with the line number for a readable script, e.g., `preview_jquery.js#652`), followed by only the core lines of the JS code that the function ports, not the surrounding code.
- Data the page serves but hides is fair game, e.g., KIN8 hides the player after a movie's streaming period ends but still sets `og:video`, so the provider uses it for the sample video. Only guessed or probed values are off-limits.
- If there is no data source and no frontend code to back a value, leave the field empty.
- If the frontend only uses HLS for the sample video, return the HLS URL and leave the MP4 URL empty.
- When only part of what a ported function returns belongs in a field, keep the port intact and filter at the call site, e.g., HEYZO `galleryImages` returns images and thumbnails, and only the images go to `PreviewImages`.
- Plain URL constants, e.g., page and API URLs, need no comments or JS code unless something about them is not obvious.

Example of a ported derivation, `provider/fanza/fanza.go`:

```go
// PreviewSrc maximizes the preview image.
// Ref: https://digstatic.dmm.com/js/digital/preview_jquery.js#652
// JS Code:
// // 画像パスの正規化
// function preview_src(src)
//
//  {
//       if (src.match(/(p[a-z]\.)jpg/)) {
//          return src.replace(RegExp.$1, 'pl.');
//       } else if (src.match(/consumer_game/)) {
//          return src.replace('js-','-');
//       } else if (src.match(/js\-([0-9]+)\.jpg$/)) {
//          return src.replace('js-','jp-');
//       } else if (src.match(/ts\-([0-9]+)\.jpg$/)) {
//          return src.replace('ts-','tl-');
//       } else if (src.match(/(\-[0-9]+\.)jpg$/)) {
//          return src.replace(RegExp.$1, 'jp' + RegExp.$1);
//       } else {
//          return src.replace('-','jp-');
//       }
//  }
func PreviewSrc(s string) string {
	if m := regexp.MustCompile(`(p[a-z]\.)jpg`).FindStringSubmatch(s); m != nil {
		return strings.Replace(s, m[1], "pl.", 1)
	} else if strings.Contains(s, "consumer_game") {
		return strings.Replace(s, "js-", "-", 1)
	} else if regexp.MustCompile(`js-\d+\.jpg$`).MatchString(s) {
		return strings.Replace(s, "js-", "jp-", 1)
	} else if regexp.MustCompile(`ts-\d+\.jpg$`).MatchString(s) {
		return strings.Replace(s, "ts-", "tl-", 1)
	} else if m := regexp.MustCompile(`(-\d+\.)jpg$`).FindStringSubmatch(s); m != nil {
		return strings.Replace(s, m[1], "jp"+m[1], 1)
	}
	return strings.Replace(s, "-", "jp-", 1)
}
```

### Layout

- Standalone (non-method) functions and their constants go at the bottom of the provider file, right above `func init()`.

## Code

- Before hand-writing a generic helper, e.g., a set built from `map[string]struct{}` for deduplication, look for one in `common/` and `collection/`, e.g., `sets.NewOrderedSet`, and use it.
- Write a value used in only one place, e.g., a timeout, inline: `context.WithTimeout(ctx, 5*time.Second)`. Name it as a constant only when it is reused.
- Don't change production code just for tests, e.g., by turning a constant into a variable that tests override. Let the test take the real duration instead; a few extra seconds are fine.
- Test only what the change needs; don't over-test.

## Comments

Comments are there to make the code easier to understand, so keep them short and few. The `Ref:` comments of ported frontend code are the exception.

- Doc comments are one or two lines on what the thing does, e.g., `// RunVM calls fn with a new VM, interrupting the scripts it runs once ctx is done.` No examples, no edge-case details, nothing the code already says.
- Short comments that help are welcome, e.g., `payloads map[string]string // filename -> RSC payload`, `return // truncated payload`, or a one-line note on what a test case checks.
- Shorten a long comment rather than deleting it; delete only comments that add nothing.

## Git

- Don't credit AI tools in commits or pull requests: no `Co-Authored-By` trailer for an AI in commit messages, and no "Generated with ..." line in PR descriptions.

### Commits

- Title commits and PRs `Type(scope): summary`, e.g., ``Fix(provider): update `KIN8` for redesigned site``. Types are `Build`, `CI`, `Chore`, `Feature`, `Fix`, `Refactor`, and `Revert`. The scope is the area changed, e.g., `provider/fanza` or `common`, and may be dropped for repo-wide changes. Wrap code and provider names in backticks.

### Pull requests

- Keep PR descriptions concise: clearly say what the PR does and roughly how, with no code snippets and no detailed reasoning. Skip the how when the change is simple. Use these sections:
  - `## Summary`: start with the related issue or PR, if any. Use `Fixes #N` (for a bug) or `Closes #N` only when the PR fully resolves the issue, since GitHub closes it on merge; otherwise use `Related to #N` or `Follows #N`. Then say briefly what the PR does and why (motivation, trade-offs). Give examples only when the case is too complex without one.
  - `## Key Changes`: bullets listing the main changes. Omit it when the Summary already covers them.
  - `## Testing`: bullets listing the tests that ran, e.g., `go test ./provider/kin8/...`. Omit it when no tests ran.
