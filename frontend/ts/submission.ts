const hs = require("highlight.js/lib/core.js");
import "regenerator-runtime/runtime";

hs.registerLanguage("cpp", require("highlight.js/lib/languages/cpp"));
hs.registerLanguage("python", require("highlight.js/lib/languages/python"));
hs.registerLanguage("java", require("highlight.js/lib/languages/java"));
hs.registerLanguage("rust", require("highlight.js/lib/languages/rust"));
hs.registerLanguage("go", require("highlight.js/lib/languages/go"));
hs.registerLanguage("pascal", require("highlight.js/lib/languages/delphi"));

hs.initHighlightingOnLoad();

// Live updating verdicts
(() => {
    type Result =
        | {
              verdict: "..." | "Compile Error";
          }
        | {
              verdict: "Scored" | "Accepted";
              score: number;
              penalty: number;
          };

    async function fetchResult(id: string): Promise<Result> {
        return (
            await fetch(`/admin/submissions/${id}/verdict`)
        ).json() as Promise<Result>;
    }

    async function fetchResultAsUser(id: string): Promise<Result> {
        return (
            await fetch(
                `/contests/${
                    (window as any).contestId
                }/submissions/${id}/verdict`,
            )
        ).json() as Promise<Result>;
    }

    for (const field of document.getElementsByClassName("live-update")) {
        const id = field.getAttribute("data-id");
        if (id) {
            const interval = setInterval(async () => {
                const result = await (field.classList.contains("as-user")
                    ? fetchResultAsUser
                    : fetchResult)(id);
                if (result.verdict !== "...") {
                    let output = result.verdict;
                    if ("score" in result) {
                        output += ` [${result.score.toFixed(2)}`;
                        if (result.penalty > 0) {
                            output += `<span class="text-gray-600"> (+${result.penalty})</span>]`;
                        } else {
                            output += "]";
                        }
                    }
                    field.innerHTML = output;
                    clearInterval(interval);
                    // If the field wants a refresh do it
                    if (field.classList.contains("refresh-on-found")) {
                        window.location.reload();
                    }
                }
            }, 1000);
        }
    }
})();
