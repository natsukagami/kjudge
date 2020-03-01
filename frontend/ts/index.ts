// Muli font
import "typeface-muli";

// require-confirm forms
(function () {
    for (const elem of document.getElementsByClassName("require-confirm")) {
        (elem as HTMLFormElement).addEventListener("submit", ev => {
            if (!confirm("Are you sure you want to delete this item?"))
                ev.preventDefault();
        })
    }
})();
