// Muli font
import "@fontsource/mulish";
import "@fontsource/ibm-plex-mono";

// Moment.js
import hd from "humanize-duration";

// Set localStorage version.
(() => {
    const versionKey = "kjudge-localstorage-version";
    const version = "1";
    localStorage.setItem(versionKey, version);
})();

// Set timezone
(function () {
    setInterval(() => {
        // Parse and update the "the current time is" nodes.
        const now = new Date();
        const nowStr = now.toUTCString();
        const iso = now.toISOString();
        const html = `${nowStr.substring(
            0,
            nowStr.length - 7,
        )} (<span class="font-mono">${iso.substring(0, iso.length - 8)}</span>)`; // Strip timezone and seconds branch
        for (const elem of document.getElementsByClassName(
            "utc-current-time",
        )) {
            const e = elem as HTMLSpanElement;
            if (e.innerHTML !== html) e.innerHTML = html;
        }
    }, 1000);

    for (const elem of document.getElementsByClassName("display-time")) {
        // Special nodes that takes a time and formats it into local time.
        const time = new Date(elem.getAttribute("data-time") || 0);
        elem.innerHTML = time.toLocaleString() + " (local)";
        elem.setAttribute("title", "UTC: " + time.toUTCString());
    }
})();

// require-confirm forms
(function () {
    for (const elem of document.getElementsByClassName("require-confirm")) {
        (elem as HTMLFormElement).addEventListener("submit", (ev) => {
            if (!confirm("Are you sure you want to delete this item?"))
                ev.preventDefault();
        });
    }
})();

// Handle login button: set href = login address + back-ref
(() => {
    for (const link of document.getElementsByClassName("login-button")) {
        (link as HTMLAnchorElement).href =
            "/user/login?last=" + encodeURIComponent(document.URL);
    }
    for (const link of document.getElementsByClassName("logout-button")) {
        (link as HTMLAnchorElement).href =
            "/user/logout?last=" + encodeURIComponent(document.URL);
    }
})();

// Handle timers
(() => {
    for (const t of document.getElementsByClassName("timer")) {
        const timer = t as HTMLDivElement;
        const start = timer.getAttribute("data-start");
        const end = timer.getAttribute("data-end");
        if (!start || !end) {
            continue;
        }

        const startTime = new Date(start);
        const endTime = new Date(end);

        const update = () => {
            const now = new Date();
            if (now.getTime() < startTime.getTime()) {
                timer.innerHTML = `Contest starting in <span class="font-semibold">${hd(
                    startTime.getTime() - now.getTime(),
                    {
                        largest: 2,
                        units: ["h", "m", "s"],
                        round: true,
                    },
                )}</span>`;
            } else if (now.getTime() < endTime.getTime()) {
                timer.innerHTML = `Time remaining: <span class="font-semibold">${hd(
                    endTime.getTime() - now.getTime(),
                    {
                        largest: 2,
                        units: ["h", "m", "s"],
                        round: true,
                    },
                )}</span>`;
            } else {
                timer.innerHTML = "Contest has ended.";
                clearInterval(interval);
            }
        };

        const interval = setInterval(update, 1000);
        update();
    }
})();

(() => {
    for (const p of document.getElementsByClassName("make-embed")) {
        p.innerHTML = `<embed src="${p.getAttribute(
            "data-src",
        )}" type="application/pdf" class="w-full" style="height: 75vh;"/>`;
    }
})();

// Handle "current-url" input fields
(() => {
    for (const input of document.getElementsByClassName("current-url")) {
        (input as HTMLInputElement).value = document.URL;
    }
})();

// Links going back
(() => {
    for (const item of document.getElementsByClassName("link-back")) {
        const link = item as HTMLAnchorElement;
        link.href = document.referrer;
        link.addEventListener("click", (e) => {
            e.preventDefault();
            history.back(); // don't push the current page into history
        });
    }
})();
