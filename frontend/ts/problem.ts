// Tab handling
(() => {
    let currentTab = new URL(document.URL).hash.substr(1);
    if (
        currentTab !== "statements" &&
        currentTab !== "files" &&
        currentTab !== "submit" &&
        currentTab !== "submissions"
    ) {
        currentTab = "statements";
    }
    const tabs = document.getElementsByClassName("tab");
    const tabButtons = document.getElementsByClassName("problem-tab");

    const render = (tab: string | null) => {
        if (!tab) return;
        for (const t of tabs) {
            const elem = t as HTMLDivElement;
            if (t.getAttribute("data-tab") === tab) {
                elem.style.display = "block";
            } else {
                elem.style.display = "none";
            }
        }

        for (const t of tabButtons) {
            if (t.getAttribute("data-tab") === tab) {
                t.classList.remove("bg-gray-200", "hover:bg-gray-400");
                t.classList.add("bg-blue-400", "text-white");
            } else {
                t.classList.add("bg-gray-200", "hover:bg-gray-400");
                t.classList.remove("bg-blue-400", "text-white");
            }
        }
    };

    render(currentTab);

    for (const t of tabButtons) {
        const tab = t.getAttribute("data-tab");
        if (!tab) {
            continue;
        }
        t.addEventListener("click", _ => render(t.getAttribute("data-tab")));
    }

    window.addEventListener("hashchange", _ =>
        render(window.location.hash.substr(1)),
    );
})();
