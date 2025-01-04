import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import axios from "axios";
import styles from "@/styles/blockDetails.module.css";

export default function BlockDetails() {
  const router = useRouter();
  const { id } = router.query; // Extract block ID from the query
  const [blockDetails, setBlockDetails] = useState(null);

  useEffect(() => {
    if (id) {
      // Fetch block details based on the ID
      axios
        .post("http://localhost:8093/query/service", {
          statement: `SELECT * FROM blocks WHERE META().id = "${id}";`,
        }, {
          headers: {
            "Content-Type": "application/json",
            Authorization: "Basic " + btoa("Administrator:password"),
          },
        })
        .then((response) => {
          setBlockDetails(response.data.results[0]);
        })
        .catch((error) => console.error("Error fetching block details:", error));
    }
  }, [id]);

  return (
    <section className={styles.blockDetails}>
      <button className={styles.backButton} onClick={() => router.back()}>
        &nbsp;&nbsp;BACK&nbsp;&nbsp;
      </button>
      <br /><br />
      {blockDetails ? (
        <div>
          <pre className={styles.pre}>
            <h3><b style={{ color: "#5ca8ff" }}>{id}</b></h3><br /><br />
            {JSON.stringify(blockDetails, null, 2)}
          </pre>
        </div>
      ) : (
        <p>Loading block details...</p>
      )}
      <br />
      <button className={styles.backButton} onClick={() => router.back()}>
        &nbsp;&nbsp;BACK&nbsp;&nbsp;
      </button>
    </section>
  );
}
