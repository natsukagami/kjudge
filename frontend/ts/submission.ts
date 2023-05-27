import "regenerator-runtime/runtime";

import hs from "highlight.js/lib/core.js";
import hsCpp from "highlight.js/lib/languages/cpp";
import hsPas from "highlight.js/lib/languages/delphi";
import hsGo from "highlight.js/lib/languages/go";
import hsJava from "highlight.js/lib/languages/java";
import hsPy from "highlight.js/lib/languages/python";
import hsRust from "highlight.js/lib/languages/rust";

hs.registerLanguage("cpp", hsCpp);
hs.registerLanguage("pascal", hsPas);
hs.registerLanguage("go", hsGo);
hs.registerLanguage("java", hsJava);
hs.registerLanguage("python", hsPy);
hs.registerLanguage("rust", hsRust);

hs.highlightAll();

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

    const fetchResult = async (id: string): Promise<Result> => {
        return (await fetch(`/admin/submissions/${id}/verdict`)).json();
    };

    const fetchResultAsUser = async (id: string): Promise<Result> => {
        return (
            await fetch(
                `/contests/${document.contestId}/submissions/${id}/verdict`,
            )
        ).json();
    };

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
