// Muli font
import "typeface-muli";

// Set timezone 
(function () {
    for (const elem of document.getElementsByClassName("utc-current-time")) {
        // Parse and update the "the current time is" nodes.
        setInterval(() => {
            const now = new Date();
            const iso = now.toISOString();
            (elem as HTMLSpanElement).innerHTML = `${now.toUTCString()} (<span class="font-mono">${iso.substr(0, iso.length - 5)}</span>)`;
        }, 1000);
    }

    for (const elem of document.getElementsByClassName("display-time")) {
        // Special nodes that takes a time and formats it into local time.
        const time = new Date(elem.getAttribute("data-time"));
        elem.innerHTML = time.toLocaleString() + " (local)";
        elem.setAttribute("title", "UTC: " + time.toUTCString())

    }
})();

// require-confirm forms
(function () {
    for (const elem of document.getElementsByClassName("require-confirm")) {
        (elem as HTMLFormElement).addEventListener("submit", ev => {
            if (!confirm("Are you sure you want to delete this item?"))
                ev.preventDefault();
        })
    }
})();

// load the list of tests
(function () {
    const SHOW = "[show]";
    const HIDE = "[hide]";
    const SHOW_ALL = "[show all]";
    const HIDE_ALL = "[hide all]";

    const testTables = Array.from(document.getElementsByClassName("tests-list"))
    const toggles = Array.from(document.getElementsByClassName("toggle-tests"))
    const groups = testTables
        .reduce((m, elem) => {
            const e = elem as HTMLDivElement;
            const group = e.getAttribute("data-test-group");
            const toggle = toggles.find(t => t.getAttribute("data-test-group") == group);
            m.set(group, [e, toggle]);
            return m;
        }, new Map<string, [HTMLDivElement, Element]>());
    let opening = 0;
    const doToggle = (table: HTMLDivElement, toggle: Element, force?: boolean) => {
        const current = table.style.maxHeight === ""; // table showing?
        if (force === undefined) {
            force = !current;
        }
        if (force === current) return;
        if (force) {
            toggle.innerHTML = HIDE;
            table.style.maxHeight = "";
            ++opening;
        } else {
            toggle.innerHTML = SHOW;
            table.style.maxHeight = "0";
            --opening;
        }
        if (opening > 0) {
            allToggle.innerHTML = HIDE_ALL;
        } else {
            allToggle.innerHTML = SHOW_ALL;
        }
    }
    const items = Array.from(groups.values());
    for (const [table, toggle] of items) {
        toggle.addEventListener("click", () => doToggle(table, toggle))
    }

    const allToggle: Element = document.getElementById("toggle-all-tests");
    allToggle.addEventListener("click", ev => {
        const switchOn = allToggle.innerHTML === SHOW_ALL;
        for (const [table, toggle] of items) {
            doToggle(table, toggle, switchOn)
        }
    });
})();
