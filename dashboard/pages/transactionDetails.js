import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import axios from "axios";
import styles from "@/styles/transactionDetails.module.css";
import Link from "next/link";

export default function TransactionDetails() {
    const router = useRouter();
    const { id } = router.query; // Extract transaction ID from the query
    const [transactionDetails, setTransactionDetails] = useState(null);

    useEffect(() => {
        if (id) {
            // Fetch transaction details based on the ID
            axios
                .post("http://localhost:8093/query/service", {
                    statement: `SELECT * FROM transactions WHERE META().id = "${id}";`,
                }, {
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: "Basic " + btoa("Administrator:password"),
                    },
                })
                .then((response) => {
                    setTransactionDetails(response.data.results[0]);
                })
                .catch((error) =>
                    console.error("Error fetching transaction details:", error)
                );
        }
    }, [id]);

    return (
        <section className={styles.transactionDetails}>
            <button className={styles.backButton} onClick={() => router.back()}>
                &nbsp;&nbsp;BACK&nbsp;&nbsp;
            </button>
            <br /><br />
            {transactionDetails ? (
                <div>
                    <pre className={styles.pre}>
                        <h3><b style={{ color: "#5ca8ff" }}>{id}</b></h3><br /><br />
                        {JSON.stringify(transactionDetails, null, 2)}
                    </pre>
                </div>
            ) : (
                <p>Loading transaction details...</p>
            )}
            <br />
            <button className={styles.backButton} onClick={() => router.back()}>
                &nbsp;&nbsp;BACK&nbsp;&nbsp;
            </button>
        </section>
    );
}
