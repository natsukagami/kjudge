// Handle "show unanswered"
(() => {
    const noUnanswered = document.getElementById("no-unanswered");
    const showUnanswered = document.getElementById(
        "show-unanswered",
    ) as HTMLInputElement | null;
    if (!showUnanswered) return;

    const scan = () => {
        let hasUnanswered = false;
        for (const div of document.getElementsByClassName("clarification")) {
            if (div.getAttribute("data-answered") === "true") {
                if (showUnanswered.checked) div.classList.add("hidden");
                else {
                    div.classList.remove("hidden");
                }
            } else hasUnanswered = true;
        }
        if (!hasUnanswered && showUnanswered.checked) {
            noUnanswered?.classList.remove("hidden");
        } else {
            noUnanswered?.classList.add("hidden");
        }
    };

    showUnanswered.addEventListener("change", scan);
})();

// Handle "template response"
(() => {
    for (const elem of document.getElementsByClassName("premade")) {
        const textarea = elem.parentElement?.getElementsByClassName(
            "form-input",
        )[0] as HTMLTextAreaElement;
        const select = elem as HTMLSelectElement;
        select.addEventListener("change", () => {
            textarea.value = select.selectedOptions[0].value || textarea.value;
        });
    }
})();
