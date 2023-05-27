export {};

declare global {
    interface Document {
        contestId: number;
        announcements: {
            setLast: (x: number | string) => void;
            markUnread: () => void;
        };
    }
}

const notificationSound = new Audio(
    new URL("../sounds/notification.ogg", import.meta.url).toString(),
);

// Stores the last announcement and clarification read.
interface Store {
    contestId: number;
    lastAnnouncement: number;
    lastClarification: number;
}

// Periodically fetch announcements
document.announcements = (() => {
    const announcementKey = "kjudge-announcement-last";
    const get = () =>
        JSON.parse(localStorage.getItem(announcementKey) as string) as Store;
    const set = (x: Store) =>
        localStorage.setItem(announcementKey, JSON.stringify(x));
    // Set a default value
    localStorage.getItem(announcementKey) === null ||
    get().contestId !== document.contestId
        ? set({
              contestId: document.contestId,
              lastAnnouncement: 0,
              lastClarification: 0,
          })
        : void 0;
    // Set announcements count!!
    const messagesCounter = document.getElementById(
        "messages-counter",
    ) as HTMLDivElement;
    const originalTitle = document.title;
    const setAnnouncementCount = (x: number) => {
        if (x > 0) {
            if (
                messagesCounter.innerHTML !== "" &&
                messagesCounter.innerHTML !== x.toString()
            ) {
                notificationSound.play();
            }
            messagesCounter.classList.remove("hidden");
            document.title = `[${x}] ${originalTitle}`;
        } else {
            messagesCounter.classList.add("hidden");
            document.title = originalTitle;
        }
        messagesCounter.innerHTML = x.toString();
    };

    // Fetch announcements count
    const fetchAnnouncements = () => {
        const info = get();
        return fetch(
            `/contests/${document.contestId}/messages/unread?last_announcement=${info.lastAnnouncement}&last_clarification=${info.lastClarification}`,
        )
            .then((v) => v.json())
            .then(setAnnouncementCount);
    };

    setInterval(fetchAnnouncements, 10 * 1000);
    const firstLoad = fetchAnnouncements();

    return {
        setLast: () => {
            firstLoad.finally(() => {
                const clars = [
                    ...document.getElementsByClassName("clarification"),
                ]
                    .filter(
                        (item) =>
                            item.getAttribute("data-responded") === "true",
                    )
                    .map((item) => Number(item.getAttribute("data-id")));
                const announcements = [
                    ...document.getElementsByClassName("announcement"),
                ].map((item) => Number(item.getAttribute("data-id")));
                set({
                    contestId: document.contestId,
                    lastAnnouncement: Math.max(...announcements, 0),
                    lastClarification: Math.max(...clars, 0),
                });
                setAnnouncementCount(0);
            });
        },
        markUnread: () => {
            // Mark the unread announcements with special backgrounds
            const store = get();
            for (const item of document.getElementsByClassName(
                "announcement",
            )) {
                if (
                    Number(item.getAttribute("data-id")) >
                    store.lastAnnouncement
                ) {
                    item.classList.add("bg-green-200", "hover:bg-green-300");
                }
            }
            for (const item of document.getElementsByClassName(
                "clarification",
            )) {
                if (
                    Number(item.getAttribute("data-id")) >
                        store.lastClarification &&
                    item.getAttribute("data-responded") === "true"
                ) {
                    item.classList.add("bg-green-200", "hover:bg-green-300");
                }
            }
        },
    };
})();
