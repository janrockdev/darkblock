import moment from "moment";
import styles from "@/styles/Home.module.css";
import Link from "next/link";

function getDT(nano) {
  return nano ? moment(nano / 1000000).format("YYYY-MM-DD HH:mm:ss") : "";
}

function getDTFromNow(nano) {
  return nano ? moment(nano / 1000000).fromNow() : "";
}

export default function SearchResults({ result, searchInput, searchType }) {
  return (
    <section className={styles.searchResults}>
      {(result.searchType === 'transaction' || result.searchType === 'payload') && (
        <table className={styles.txnSection}>
          <thead>
            <tr className={styles.txnTitle}>
              <th>Transaction Detail</th>
              <th>Timestamp</th>
              <th>Address To</th>
              <th>Payload</th>
            </tr>
          </thead>
          <tbody>
            {result.result.results.map((txn) => {
              return (
                <tr key={txn.id} className={styles.txn}>
                  <td className={styles.blueText}>
                    <Link href={`/transactionDetails?id=${txn.id}`}>
                      {txn.id.slice(0, 16)}...{txn.id.slice(-6)}
                    </Link>
                  </td>
                  <td>{getDT(txn.transactions.timestamp)}</td>
                  <td>{txn.transactions.outputs[0].address.slice(0, 16)}...{txn.transactions.outputs[0].address.slice(-6)}</td>
                  <td className={styles.blueText}>
                    <Link href={`/transactionDetails?id=${txn.id}`}>
                      {txn.transactions.outputs[0].payload.slice(0, 16)}...{txn.transactions.outputs[0].payload.slice(-6)}
                    </Link>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
      {result.searchType === 'block' && (
        <table className={styles.txnSection}>
          <thead>
            <tr className={styles.txnTitle}>
              <th>#Block</th>
              <th>Timestamp</th>
              <th>Height</th>
              <th>Transactions</th>
            </tr>
          </thead>
          <tbody>
            {result.result.results.map((block) => {
              return (
                <tr key={block.id} className={styles.txn}>
                  <td className={styles.blueText}>
                    <Link href={`/blockDetails?id=${block.id}`}>
                      {block.id.slice(0, 23)}...{block.id.slice(-6)}
                    </Link>
                  </td>
                  <td>{getDT(block.blocks.header.timestamp)}</td>
                  <td>{block.blocks.header.height}</td>
                  <td>{block.blocks.transactions.length}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </section>
  );
}
