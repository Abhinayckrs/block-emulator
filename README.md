# BlockEmulator with Single Leader Consensus (SLC) 🚀

[![MIT License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/blockemulator-slc)](https://goreportcard.com/report/github.com/yourusername/blockemulator-slc)
[![Build Status](https://img.shields.io/github/actions/workflow/status/yourusername/blockemulator-slc/go.yml)](https://github.com/yourusername/blockemulator-slc/actions)

---

## Overview

**BlockEmulator** is a sharded blockchain simulation framework originally developed by Huang Sys Lab. This fork introduces **Single Leader Consensus (SLC)** — a lightweight, high-throughput intra-shard consensus protocol designed to maximize TPS by employing a single leader for transaction validation.  
SLC leverages a single-phase Proposal-Accept mechanism and reuses the Relay protocol for cross-shard communication, ignoring fault tolerance mechanisms for simplicity and performance.

---

## Features ✨

- Single leader validation per shard (leader is NodeID 0)
- One-phase Proposal-Accept consensus mechanism
- Followers accept blocks without validation (checking only digest)
- High throughput compared to PBFT-based Relay consensus
- Reuse of Relay protocol for cross-shard transactions
- Simple and performant: no view changes or fault tolerance

---

## Directory Structure 📂

    consensus_shard/pbft_all/
      ├── pbft.go                 # PBFT node logic & message handling
      ├── pbftMod_interface.go    # Intra- & inter-shard interfaces
      ├── messageHandle.go        # PBFT phases: Propose, Prepare, Commit
      ├── view_change.go          # Leader changes (ignored in SLC)
      ├── slc.go                  # SLC protocol implementation
      ├── pbftInside_module.go    # Intra-shard Relay logic
      ├── pbftOutside_module.go   # Cross-shard Relay logic

    message/
      ├── message.go              # Core messages (CProposal, CAccept)
      ├── message_relay.go        # Cross-shard Relay messages

    core/
      ├── block.go                # Block struct & operations
      ├── txpool.go               # Transaction pool management

    params/
      ├── global_config.go        # Global config parameters
      ├── static_config.go        # Static chain configs

    supervisor/committee/
      ├── committee_relay.go      # Node initialization & transaction injection

---

## Getting Started 🚀

### Prerequisites

- Go 1.19+  
- Clone this repo

### Setup & Configuration

Update `paramsConfig.json` with:

    {
      "ConsensusMethod": 4,
      "Block_Interval": 2000,
      "MaxBlockSizeGlobal": 3000
      // ... other settings ...
    }

### How to Run Your Simulator in the Terminal

1. **Modify `params/static_config.go`**  
   Add your new protocol to the `CommitteeMethod` list if it’s not already included.

2. **Update `paramsConfig.json`**  
   Adjust configuration variables as needed, such as `ConsensusMethod`, `Block_Interval`, and `MaxBlockSizeGlobal`.

3. **Run the Build Script**  
   Execute the appropriate pre-compile script in the `zPreCompileScripts` directory based on your OS:  
   - For Linux/macOS, run the `.sh` script.  
   - For Windows, run the `.bat` file.

4. **Run the Simulator Executable with Flags**  
   Use the compiled executable and specify the number of shards and nodes with flags. For example, on Linux:  
   ```bash
   ./blockEmulator_Linux_Precompile -g --shellForExe -S2 -N4
   ```
5. **Execute the Generated Script** 
   After running the above command, a new script file will be generated. Run the script to start the simulation with your configured settings.


    

Monitor logs in `expTest/log` and results in `expTest/result`.

---

## How to Implement Your Own Sharded Protocol 🛠️

1. **Define Consensus Logic**  
   Implement the `ExtraOpInConsensus` interface in a new file under `consensus_shard/pbft_all`.

2. **Add New Message Types**  
   Extend `message/message.go` with your protocol’s message types and structs.

3. **Modify `pbft.go`**  
   Register your protocol in `NewPbftNode` and handle your message types.

4. **Update Committee Initialization**  
   Modify `supervisor/committee/committee_relay.go` to initialize your protocol.

5. **Configure Parameters**  
   Update `global_config.go`, `static_config.go`, and `paramsConfig.json`.

6. **Cross-Shard Logic**  
   Reuse or implement `OpInterShards` interface as needed.

7. **Testing & Measurement**  
   Use provided CSV transactions and measurement tools .

---

## Single Leader Consensus (SLC) Details ⚡

- Leader (NodeID 0) proposes blocks (`CProposal` messages).  
- Followers send acceptance (`CAccept` messages) without validation.  
- Leader commits after receiving `(2f + 1)` accepts.  
- No prepare/commit phases or view changes.

### Performance Tips

- Lower `Block_Interval` for faster blocks (default 2000ms).  
- Increase `MaxBlockSizeGlobal` to increase transactions per block.

---



## Contributing 🤝

Feel free to fork, extend, and submit pull requests!  
Follow the guide above to implement new protocols or improve SLC.

---

## License 📄

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---
