// Batch add form confirmation
(() => {
    const batchAddForm = document.getElementById(
        "batch-add-form",
    ) as HTMLFormElement;
    const resetCheckbox = document.getElementById(
        "batch-add-reset",
    ) as HTMLInputElement;
    batchAddForm.addEventListener("submit", (e) => {
        if (
            resetCheckbox.checked &&
            !confirm(
                "Are you sure you want to continue? This will delete ALL users and ALL submissions!",
            )
        ) {
            e.preventDefault();
        }
    });
})();
