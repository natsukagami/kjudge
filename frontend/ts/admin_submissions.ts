// Submission table interactions
(() => {
    function handleTable(table: HTMLTableElement) {
        const selectAll = table.getElementsByClassName(
            "select-all",
        )[0] as HTMLTableHeaderCellElement;
        const rows = table.getElementsByClassName(
            "table-row",
        ) as HTMLCollectionOf<HTMLTableRowElement>;
        const checkboxesArray = Array.of(...rows).map(
            v =>
                v.getElementsByClassName(
                    "select-to-rejudge",
                )[0] as HTMLInputElement,
        );
        const checkboxes: Map<string, HTMLInputElement> = new Map();
        for (const chk of checkboxesArray) {
            const id = chk.getAttribute("value");
            if (!id) continue;
            checkboxes.set(id, chk);
        }

        // the filters
        const userFilter = table
            .getElementsByClassName("user-filter")
            .item(0) as HTMLSelectElement | null;
        const problemFilter = table
            .getElementsByClassName("problem-filter")
            .item(0) as HTMLSelectElement | null;

        const selectedInputs = table.getElementsByClassName(
            "selected-submissions",
        ) as HTMLCollectionOf<HTMLInputElement>;
        const selectedCount = table.getElementsByClassName(
            "selected-count",
        )[0] as HTMLSpanElement;

        function updateInputs() {
            const checked: string[] = [];
            for (const u of checkboxesArray) {
                const row = u.parentElement?.parentElement;
                if (u.checked && row && row.style.display !== "none") {
                    checked.push(u.value);
                }
            }
            for (const u of selectedInputs) {
                u.value = checked.join(",");
            }
            if (checked.length > 0) {
                selectedCount.innerHTML = `(${checked.length})`;
            } else {
                selectedCount.innerHTML = "";
            }
        }

        function filterRows(): [HTMLTableRowElement[], HTMLTableRowElement[]] {
            const yes = [],
                no = [];
            const userID =
                userFilter?.options[userFilter.selectedIndex].value ?? "";
            const problemIDs = (
                problemFilter?.options[
                    problemFilter.selectedIndex
                ].value?.split(",") ?? []
            ).filter(v => v !== "");
            for (const row of rows) {
                if (
                    (userID === "" ||
                        row.getAttribute("data-user") === userID) &&
                    (problemIDs.length === 0 ||
                        problemIDs.indexOf(
                            row.getAttribute("data-problem") ?? "",
                        ) !== -1)
                ) {
                    yes.push(row);
                } else {
                    no.push(row);
                }
            }
            return [yes, no];
        }

        function setChecked(
            yes: HTMLTableRowElement[],
            no: HTMLTableRowElement[],
        ) {
            for (const r of yes) {
                const id = r.getAttribute("data-id") ?? "";
                const v = checkboxes.get(id);
                if (v) {
                    v.checked = true;
                }
            }
            for (const r of no) {
                const id = r.getAttribute("data-id") ?? "";
                const v = checkboxes.get(id);
                if (v) {
                    v.checked = false;
                }
            }
        }

        function setVisible(
            yes: HTMLTableRowElement[],
            no: HTMLTableRowElement[],
        ) {
            for (const row of yes) {
                row.style.display = "";
            }
            for (const row of no) {
                row.style.display = "none";
            }
        }

        // Handle select-all
        selectAll.addEventListener("click", () => {
            const [yes, no] = filterRows();
            if (selectedInputs[0].value === "") {
                // SELECT all
                setChecked(yes, no);
            } else {
                // DESELECT all
                setChecked([], [...yes, ...no]);
            }
            updateInputs();
        });

        // Handle filters
        userFilter?.addEventListener("change", _ => {
            const [yes, no] = filterRows();
            setVisible(yes, no);
            updateInputs();
        });
        problemFilter?.addEventListener("change", _ => {
            const [yes, no] = filterRows();
            setVisible(yes, no);
            updateInputs();
        });

        // Handle checkboxes
        for (const c of checkboxesArray) {
            c.addEventListener("click", () => updateInputs());
        }

        // Run a single round of update
        (() => {
            const [yes, no] = filterRows();
            setVisible(yes, no);
            updateInputs();
        })();
    }

    const tables = document.getElementsByClassName("submissions-table");
    for (const table of tables) {
        handleTable(table as HTMLTableElement);
    }
})();

import "regenerator-runtime/runtime";

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
                        output += ` [${result.score.toFixed(2)} | ${
                            result.penalty
                        }]`;
                    }
                    field.innerHTML = output;
                    clearInterval(interval);
                }
            }, 1000);
        }
    }
})();
