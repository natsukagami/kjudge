import hs from "highlight.js";

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

    for (const field of document.getElementsByClassName("live-update")) {
        const id = field.getAttribute("data-id");
        if (id) {
            const interval = setInterval(async () => {
                const result = await fetchResult(id);
                if (result.verdict !== "...") {
                    let output = result.verdict;
                    if ("score" in result) {
                        output += ` [${result.score.toFixed(2)}`;
                        if (result.penalty > 0) {
                            output += `<span class="text-gray-800"> (+${result.penalty})</span>]`;
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
