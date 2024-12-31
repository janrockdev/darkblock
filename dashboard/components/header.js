import { useEffect, useState } from "react";
import axios from "axios";
import Image from "next/image";
import styles from "@/styles/Home.module.css";

export default function Header() {
  const [ethPrice, setEthPrice] = useState("");

  useEffect(() => {
    const getEthPrice = async () => {
      // const response = await axios.get("http://localhost:5001/getethprice", {});
      setEthPrice(1);
    };
    getEthPrice();
  });

  return (
    <section className={styles.header}>
      <section className={styles.topHeader}>
        Environment:{" "}
        <span className={styles.blueText}>&nbsp;DEVNET</span>
      </section>
      <section className={styles.navbar}>
        DARKBLOCK SCAN
      </section>
    </section>
  );
}
