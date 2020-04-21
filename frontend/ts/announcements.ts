declare interface Window {
    contestId: number;
    announcements: {
        setLast: (x: number | string) => void;
        markUnread: () => void;
    };
}

const notificationSound = new Audio(require("../sounds/notification.ogg"));

// Periodically fetch announcements
window.announcements = (() => {
    const announcementKey = "kjudge-announcement-last";
    const get = () => Number(localStorage.getItem(announcementKey) as string);
    const set = (x: number | string) =>
        localStorage.setItem(announcementKey, x.toString());
    // Set a default value
    localStorage.getItem(announcementKey) === null ? set(0) : void 0;
    // Set announcements count!!
    const announcementCounter = document.getElementById(
        "announcement-counter",
    ) as HTMLDivElement;
    const originalTitle = document.title;
    const setAnnouncementCount = (x: number) => {
        if (x > 0) {
            if (
                announcementCounter.innerHTML !== "" &&
                announcementCounter.innerHTML !== x.toString()
            ) {
                notificationSound.play();
            }
            announcementCounter.classList.remove("hidden");
            document.title = `[${x}] ${originalTitle}`;
        } else {
            announcementCounter.classList.add("hidden");
            document.title = originalTitle;
        }
        announcementCounter.innerHTML = x.toString();
    };

    // Fetch announcements count
    const fetchAnnouncements = () => {
        return fetch(
            `/contests/${window.contestId}/announcements/unread?since=${get()}`,
        )
            .then((v) => v.json())
            .then(setAnnouncementCount);
    };

    setInterval(fetchAnnouncements, 10 * 1000);
    const firstLoad = fetchAnnouncements();

    return {
        setLast: (x: number | string) => {
            firstLoad.then(() => {
                set(x);
                setAnnouncementCount(0);
            });
        },
        markUnread: () => {
            // Mark the unread announcements with special backgrounds
            const lastRead = get();
            for (const item of document.getElementsByClassName(
                "announcement",
            )) {
                if (Number(item.getAttribute("data-id")) > lastRead) {
                    item.classList.add("bg-green-200", "hover:bg-green-300");
                }
            }
        },
    };
})();
