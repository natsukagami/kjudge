import { h, render } from "preact";
import { useState } from "preact/hooks";
import FlipMove from "react-flip-move";
import "regenerator-runtime/runtime";
import { useEffect } from "react";

declare global {
    interface Document {
        initialScoreboard: Scoreboard;
    }
}

interface Scoreboard {
    contest_id: number;
    contest_type: "weighted" | "unweighted";
    problems: Problem[];
    users: User[];
}

interface Problem {
    id: number;
    name: string;
    display_name: string;
}

interface User {
    id: string;
    rank: number;
    total_penalty: number;
    solved_problems: number;
    total_score: number;
    problem_results: { [key: number]: ProblemResult };
}

interface ProblemResult {
    score: number;
    solved: boolean;
    penalty: number;
    failed_attempts: number;
}

/**
 * Updates the scoreboard, given the old one.
 */
async function updateScoreboard(old: Scoreboard): Promise<Scoreboard> {
    return (await fetch(`/contests/${old.contest_id}/scoreboard/json`)).json();
}

/**
 * The main scoreboard component.
 * */
const App = ({ initScoreboard }: { initScoreboard: Scoreboard }) => {
    const [[scoreboard, lastUpdated], update] = useState([
        initScoreboard,
        new Date(),
    ]);
    const fetchScoreboard = async () => {
        try {
            const sb = await updateScoreboard(scoreboard);
            update([sb, new Date()]);
        } catch (e) {
            console.error(`Scoreboard update failed: ${e}`);
        }
    };
    useEffect(() => {
        const interval = setInterval(fetchScoreboard, 5000);
        return () => clearInterval(interval);
    }, []);
    return (
        <div class="w-full">
            <Headers {...scoreboard} />
            <FlipMove>
                {scoreboard.users.map((u) => (
                    <div key={u.id}>
                        <Row key={u.id} user={u} {...scoreboard} />
                    </div>
                ))}
            </FlipMove>
            <div class="my-2 text-right px-2 text-gray-800">
                <span>Last updated: {lastUpdated.toString()}. </span>
                <a
                    href="#"
                    class="hover:text-green-600"
                    onClick={fetchScoreboard}
                >
                    [update]
                </a>
            </div>
        </div>
    );
};

/**
 * Take a list of problems and produce the table headers.
 */
const Headers = ({
    contest_id,
    problems,
}: Pick<Scoreboard, "contest_id" | "problems">) => {
    return (
        <div class="flex flex-row items-stretch">
            <div
                class="text-lg py-1 border-b flex-table-cell text-center flex-shrink-0"
                style="width: 4rem;"
            >
                Rank
            </div>
            <div class="text-lg py-1 flex-table-cell border-b border-l pl-8 text-left flex-grow">
                Username
            </div>
            <div
                class="text-lg py-1 flex-table-cell border-b border-l text-center px-1 flex-shrink-0"
                style="width: 6rem;"
            >
                Total Score
            </div>
            {problems.map((p) => (
                <div
                    key={p.id}
                    class="text-lg py-1 flex-table-cell border-b border-l text-center px-1 flex-shrink-0"
                    style="width: 4rem;"
                >
                    <a
                        href={`/contests/${contest_id}/problems/${p.name}#statements`}
                        title={`${p.name}. ${p.display_name}`}
                        class="cursor-pointer hover:text-blue-600"
                    >
                        {p.name}
                    </a>
                </div>
            ))}
        </div>
    );
};

/**
 * Row renders an user row.
 */
const Row = ({
    contest_type,
    problems,
    user,
}: {
    contest_type: Scoreboard["contest_type"];
    problems: Scoreboard["problems"];
    user: User;
}) => {
    const totalScore =
        contest_type === "unweighted" ? user.solved_problems : user.total_score;
    return (
        <div class="flex flex-row items-stretch w-full">
            <div
                class="text-lg py-3 border-b text-center flex-table-cell flex-shrink-0"
                style="width: 4rem;"
            >
                {user.rank}
            </div>
            <div class="text-lg py-3 border-b pl-8 border-l flex-table-cell flex-grow">
                {user.id}
            </div>
            <div
                class="text-lg py-3 border-b text-center font-semibold border-l flex-table-cell flex-shrink-0"
                style="width: 6rem;"
            >
                <div>{totalScore}</div>
                {user.total_penalty > 0 ? (
                    <div class="text-sm text-gray-600">
                        {user.total_penalty}
                    </div>
                ) : null}
            </div>
            {problems.map((p) => (
                <Cell
                    key={p.id}
                    contest_type={contest_type}
                    result={user.problem_results[p.id]}
                ></Cell>
            ))}
        </div>
    );
};

/**
 * Return a cell containing the result of a problem given the contest_type and its result.
 */
const Cell = ({
    contest_type,
    result,
}: {
    contest_type: Scoreboard["contest_type"];
    result: ProblemResult;
}) => {
    let score: string = "";
    let color_class: string = "";
    let title: string = "";

    if (contest_type === "unweighted") {
        if (result.solved) {
            score = `+${
                result.failed_attempts > 0 ? result.failed_attempts : ""
            }`;
            color_class = "text-green-700";
        } else if (result.failed_attempts > 0) {
            score = `-${result.failed_attempts}`;
            color_class = "text-red-600";
        } else {
            score = "-";
        }
    } else {
        score = `${result.score}`;
        if (result.solved) {
            color_class = "text-green-700";
            title = `${result.failed_attempts + 1} attempts`;
        } else if (result.score > 0) {
            color_class = "text-orange-400";
            title = `${result.failed_attempts} attempts`;
        } else if (result.failed_attempts > 0) {
            color_class = "text-red-600";
            title = `${result.failed_attempts} attempts`;
        } else {
            score = "-";
        }
    }

    return (
        <div
            class={
                "py-3 border-b font-semibold text-center border-l flex-table-cell flex-shrink-0 " +
                color_class
            }
            style="width: 4rem;"
        >
            <div class="text-lg" title={title}>
                {score}
            </div>
            {result.penalty > 0 ? (
                <div class="font-normal text-sm text-gray-600">
                    {result.penalty}
                </div>
            ) : null}
        </div>
    );
};

(() => {
    const elem = document.getElementById("scoreboard");
    if (elem) render(<App initScoreboard={document.initialScoreboard} />, elem);
})();
