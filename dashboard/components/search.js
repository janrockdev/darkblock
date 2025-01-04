import { useState } from "react";
import axios from "axios";
import styles from "@/styles/Home.module.css";

import SearchResults from "./searchResults.js";

export default function Search() {
  const [showResult, setShowResult] = useState(false);
  const [result, setResult] = useState([]);
  const [searchInput, setSearchInput] = useState("");
  const [searchType, setSearchType] = useState(""); // New state variable for search type

  const changeHandler = (e) => {
    setSearchInput(e.target.value);
  };

  const handleSearch = async () => {
    const searchInput = document.querySelector("#inputField").value;
    document.querySelector("#inputField").value = "";

    let query = '';
    let type = '';

    if (searchInput.startsWith('tx::')) {
      query = `SELECT META().id AS id, * FROM transactions WHERE META().id = "${searchInput}";`;
      type = 'transaction';
    } else if (searchInput.startsWith('block::')) {
      query = `SELECT META().id AS id, * FROM blocks WHERE META().id = "${searchInput}";`;
      type = 'block';
    } else if (searchInput.startsWith('ey')) {
      query = `SELECT META().id AS id, * FROM transactions WHERE outputs[0].payload = "${searchInput}";`;
      type = 'payload';
    } else {
      // Handle invalid input or default case
      console.error('Invalid input format');
      return;
    }

    const response = await axios.post("http://localhost:8093/query/service", {
      statement: query,
    }, {
      headers: {
        'Content-Type': 'application/json',
        // Add authentication headers if needed
        'Authorization': 'Basic ' + btoa('Administrator:password')
      }
    });

    console.log(response.data);
    setResult(response.data);
    setSearchType(type); // Set the search type
    setShowResult(true);
  };

  return (
    <section className={styles.searchContainer}>
      <section className={styles.searchHeader}>
        <section className={styles.searchSection}>
          {/* <h3>The DarkBlock Blockchain Explorer</h3> */}
          <section className={styles.input_section}>
            <input
              className={styles.inputField}
              type="text"
              id="inputField"
              name="inputField"
              maxLength="160"
              placeholder="Search by Txn Hash / Block Hash / Payload"
              required
              onChange={changeHandler}
            />
            <button className={styles.btn} onClick={handleSearch}>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="currentColor"
                className={styles.magnifying}
              >
                <path
                  fillRule="evenodd"
                  d="M10.5 3.75a6.75 6.75 0 100 13.5 6.75 6.75 0 000-13.5zM2.25 10.5a8.25 8.25 0 1114.59 5.28l4.69 4.69a.75.75 0 11-1.06 1.06l-4.69-4.69A8.25 8.25 0 012.25 10.5z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          </section>
          <section className={styles.sponsored}>
            Beta Version:{" "}
            Data can be removed at any time.{" "}
            {/* <span className={styles.claim}>Release Plan!</span> */}
          </section>
        </section>
      </section>
      {showResult && <SearchResults result={{ result, searchInput, searchType }} />}
    </section>
  );
}