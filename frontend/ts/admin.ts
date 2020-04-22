// Unanswered clarifications
(() => {
    const counter = document.getElementById("unanswered-counter");
    if (!counter) return;
    const update = () => {
        fetch("/admin/clarifications?unanswered=true")
            .then((res) => res.json())
            .then((count: number) => {
                if (count > 0) {
                    counter.classList.remove("hidden");
                    counter.innerHTML = count.toString();
                } else {
                    counter.classList.add("hidden");
                }
            });
    };
    setInterval(update, 10 * 1000);
    update();
})();
