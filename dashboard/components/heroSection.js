import { useEffect, useState } from "react";
import Link from "next/link";
import axios from "axios";
import moment from "moment";
import styles from "@/styles/Home.module.css";
import {
  faCube,
  faGauge,
  faGlobe,
  faServer,
  faFileContract,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useRouter } from 'next/router';

function getDT(nano) {
  return nano ? moment(nano / 1000000).format("YYYY-MM-DD HH:mm:ss") : "";
}

function getDTFromNow(nano) {
  return nano ? moment(nano / 1000000).fromNow() : "";
}

function convertHash(encodedAddress) {
  // Decode the base64-encoded address into bytes
  const decodedBytes = Buffer.from(encodedAddress, 'base64');

  // You may need to apply further formatting here depending on the blockchain
  // For example, you can directly return the decoded bytes as a hexadecimal string
  const blockchainAddress = decodedBytes.toString();

  return blockchainAddress;
}

export default function HeroSection() {
  const [showResult, setShowResult] = useState(true);
  const [blockResult, setBlockResult] = useState([]);
  const [transactionsResult, setTransactionsResult] = useState([]);
  const [totalTransactions, setTotalTransactions] = useState("");
  const [totalBlocks, setTotalBlocks] = useState("");
  const [blockIds, setBlockIds] = useState([]);
  const [txIds, setTxsIds] = useState([]);

  const router = useRouter();
  const handleSearchRedirect = (id, type) => {
    const prefix = type === 'block' ? 'block::' : 'tx::';
    router.push(`/search?query=${id}`);
  };

  useEffect(() => {
    const getTotalTransactions = async () => {
      try {
        const response = await axios.post('http://localhost:8093/query/service', {
          statement: 'SELECT count(*) as response FROM `transactions`;',
        }, {
          headers: {
            'Content-Type': 'application/json',
            // Add authentication headers if needed
            'Authorization': 'Basic ' + btoa('Administrator:password')
          }
        });
        setTotalTransactions(response.data.results[0].response);
      } catch (error) {
        console.error("Error querying Couchbase:", error);
      }
    };

    const getTotalBlocks = async () => {
      try {
        const response = await axios.post('http://localhost:8093/query/service', {
          statement: 'SELECT count(*) as response FROM `blocks`;',
        }, {
          headers: {
            'Content-Type': 'application/json',
            // Add authentication headers if needed
            'Authorization': 'Basic ' + btoa('Administrator:password')
          }
        });
        setTotalBlocks(response.data.results[0].response);
      } catch (error) {
        console.error("Error querying Couchbase:", error);
      }
    };

    const getBlocksIds = async () => {
      try {
        const response = await axios.post('http://localhost:8093/query/service', {
          statement: 'SELECT META().id AS id, header.timestamp AS timestamp FROM `blocks` ORDER BY META().id DESC LIMIT 10;',
        }, {
          headers: {
            'Content-Type': 'application/json',
            // Add authentication headers if needed
            'Authorization': 'Basic ' + btoa('Administrator:password')
          }
        });
        setBlockIds(response.data.results) //.map((block) => block.id));
      } catch (error) {
        console.error("Error querying Couchbase:", error);
      }
    };

    const getTxsIds = async () => {
      try {
        const response = await axios.post('http://localhost:8093/query/service', {
          statement: 'SELECT META().id AS id, timestamp, inputs[0].prevTxHash as orig FROM `transactions` ORDER BY timestamp DESC LIMIT 10;',
        }, {
          headers: {
            'Content-Type': 'application/json',
            // Add authentication headers if needed
            'Authorization': 'Basic ' + btoa('Administrator:password')
          }
        });
        setTxsIds(response.data.results) //.map((block) => block.id));
      } catch (error) {
        console.error("Error querying Couchbase:", error);
      }
    };

    getTotalTransactions();
    getTotalBlocks();
    getBlocksIds();
    getTxsIds();

    const intervalId = setInterval(() => {
      getTotalTransactions();
      getTotalBlocks();
      getBlocksIds();
      getTxsIds();
    }, 5000);

    return () => clearInterval(intervalId);
  }, []);

  return (
    <section className={styles.heroSectionContainer}>
      {showResult && (
        <section>
          <section className={styles.latestResults_header}>
            <section>
              <section className={styles.latestResults_box}>
                <section className={styles.svgSection}>
                  <FontAwesomeIcon icon={faGlobe} className={styles.svgIcons} />
                </section>
                <section className={styles.hero_box}>
                  <p>CONNECTED NODES</p>
                  <p className={styles.heroValues}>
                    3
                  </p>
                </section>
              </section>
              <span className={styles.divider}></span>
              <section className={styles.latestResults_box}>
                <section className={styles.svgSection}>
                  <FontAwesomeIcon icon={faGlobe} className={styles.svgIcons} />
                </section>
                <section className={styles.hero_box}>
                  <p>VALIDATORS</p>
                  <p className={styles.heroValues}>1</p>
                </section>
              </section>
            </section>
            <section>
              <section className={styles.latestResults_box}>
                <section className={styles.svgSection}>
                  <FontAwesomeIcon
                    icon={faServer}
                    className={styles.svgIcons}
                  />
                </section>
                <section className={styles.hero_box}>
                  <p>TRANSACTIONS</p>
                  <p className={styles.heroValues}>{totalTransactions}</p>
                </section>
              </section>
              <span className={styles.divider}></span>
              <section className={styles.latestResults_box}>
                <section className={styles.svgSection}>
                  <FontAwesomeIcon icon={faServer} className={styles.svgIcons} />
                </section>
                <section className={styles.hero_box}>
                  <p>BLOCKS</p>
                  <p className={styles.heroValues}>{totalBlocks}</p>
                </section>
              </section>
            </section>
          </section>
          <section className={styles.latestResults_body}>
            <section>
              <section className={styles.latestResults_body_title}>
                Latest Blocks
              </section>
              <table className={styles.latestResults_body_table}>
                <tbody>
                  {blockIds.map((block) => (
                    <tr
                      className={`${styles.latestResults_body_tr} ${blockResult.indexOf(block) === blockResult.length - 1 && styles.lastTd
                        }`}
                      key={block.id}
                    >
                      <td className={styles.tdIcon}>
                        <FontAwesomeIcon icon={faCube} />
                      </td>
                      <td className={styles.tdBlock}>
                        <section className={styles.blueText}>
                          <Link href={`/blockDetails?id=${block.id}`} className={styles.blueText}>
                            {block.id.split('_')[0]}_{block.id.split('_')[1].slice(0, 3)}...
                          </Link>
                        </section>
                        <section>
                          {getDT(block.timestamp)}&nbsp;({getDTFromNow(block.timestamp)})
                        </section>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
            <section>
              <section className={styles.latestResults_body_title}>
                Latest Transactions
              </section>
              <table className={styles.latestResults_body_table}>
                <tbody>
                  {txIds.map((txn) => (
                    <tr
                      className={`${styles.latestResults_body_tr} ${transactionsResult.indexOf(txn) === transactionsResult.length - 1 && styles.lastTd
                        }`}
                      key={txn.id}
                    >
                      <td className={styles.tdContract}>
                        <FontAwesomeIcon icon={faFileContract} className={styles.tdContract} />
                      </td>
                      <td className={styles.tdBlock}>
                        <section className={styles.blueText}>
                          <Link href={`/transactionDetails?id=${txn.id}`} className={styles.blueText}>
                            {txn.id.slice(0, 7)} (tx::{convertHash(txn.orig)?.slice(0, 3)})
                          </Link>
                        </section>
                        <section>
                          {getDT(txn.timestamp)}&nbsp;({getDTFromNow(txn.timestamp)})
                        </section>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          </section>
        </section>
      )}
    </section>
  );
}