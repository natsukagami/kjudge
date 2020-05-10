import hd from "humanize-duration";

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
        t.addEventListener("click", (_) => render(t.getAttribute("data-tab")));
    }

    window.addEventListener("hashchange", (_) =>
        render(window.location.hash.substr(1)),
    );
})();

// Submit form seconds-between-submissions
(() => {
    const submitForm = document.getElementById(
        "submit-form",
    ) as HTMLFormElement;
    const dataDiv = submitForm.getElementsByClassName("data")[0] as
        | HTMLDivElement
        | undefined;
    // Only continue if we have a data-div
    if (!dataDiv) return;

    const lastSubmissionTime = new Date(
        dataDiv.getAttribute("data-last-submission-time") ?? "",
    );
    const secondsBetweenSubmissions = Number(
        dataDiv.getAttribute("data-seconds-between-submissions"),
    );
    const nextSubmitTime = new Date(
        lastSubmissionTime.getTime() + secondsBetweenSubmissions * 1000,
    );

    // Set a submit handler
    submitForm.addEventListener("submit", (e) => {
        const now = new Date();
        if (now.getTime() < nextSubmitTime.getTime()) e.preventDefault();
    });

    // Set a "disable submit handler"
    const submit = submitForm.getElementsByClassName(
        "submit",
    )[0] as HTMLInputElement;
    const checkSubmit = () => {
        const now = new Date();
        if (now.getTime() < nextSubmitTime.getTime()) {
            submit.value = `Please wait ${hd(
                nextSubmitTime.getTime() - now.getTime(),
                { largest: 1, units: ["m", "s"], round: true },
            )}`;
            submit.classList.add("cursor-not-allowed");
            submit.disabled = true;
            submit.classList.remove("hover:bg-green-300");
        } else {
            submit.value = "Submit";
            submit.classList.remove("cursor-not-allowed");
            submit.disabled = false;
            submit.classList.add("hover:bg-green-300");
        }
    };
    setInterval(checkSubmit, 1000);
    checkSubmit();
})();
